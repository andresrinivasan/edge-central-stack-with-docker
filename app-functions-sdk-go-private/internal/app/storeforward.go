//
// Copyright (c) 2021 One Track Consulting
// Copyright (C) 2023 IOTech Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/redis"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/environment"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	v2Common "github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"strings"
	"sync"
	"time"
)

// RegisterCustomStoreFactory allows registration of alternative storage implementation to back the Store&Forward loop
func (svc *Service) RegisterCustomStoreFactory(name string, factory func(cfg interfaces.DatabaseInfo, cred bootstrapConfig.Credentials) (interfaces.StoreClient, error)) error {
	if name == db.RedisDB {
		return fmt.Errorf("cannot register factory for reserved name %q", name)
	}

	if svc.customStoreClientFactories == nil {
		svc.customStoreClientFactories = make(map[string]func(db interfaces.DatabaseInfo, cred bootstrapConfig.Credentials) (interfaces.StoreClient, error), 1)
	}

	svc.customStoreClientFactories[strings.ToUpper(name)] = factory

	return nil
}

func (svc *Service) createStoreClient(database interfaces.DatabaseInfo, credentials bootstrapConfig.Credentials, tlsConfig *tls.Config) (interfaces.StoreClient, error) {
	switch strings.ToLower(database.Type) {
	case db.RedisDB:
		return redis.NewClient(database, credentials, tlsConfig)
	default:
		if factory, found := svc.customStoreClientFactories[strings.ToUpper(database.Type)]; found {
			return factory(database, credentials)
		}
		return nil, db.ErrUnsupportedDatabase
	}
}

func (svc *Service) startStoreForward() {
	var storeForwardEnabledCtx context.Context
	svc.ctx.storeForwardWg = &sync.WaitGroup{}
	storeForwardEnabledCtx, svc.ctx.storeForwardCancelCtx = context.WithCancel(context.Background())
	svc.runtime.StartStoreAndForward(svc.ctx.appWg, svc.ctx.appCtx, svc.ctx.storeForwardWg, storeForwardEnabledCtx, svc.serviceKey)
}

func (svc *Service) stopStoreForward() {
	svc.LoggingClient().Info("Canceling Store and Forward retry loop")
	svc.ctx.storeForwardCancelCtx()
	svc.ctx.storeForwardWg.Wait()
}

func initializeStoreClient(config *common.ConfigurationStruct, svc *Service) error {
	// Only need the database client if Store and Forward is enabled
	if !config.Writable.StoreAndForward.Enabled {
		svc.dic.Update(di.ServiceConstructorMap{
			container.StoreClientName: func(get di.Get) interface{} {
				return nil
			},
		})
		return nil
	}

	logger := bootstrapContainer.LoggingClientFrom(svc.dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(svc.dic.Get)

	var err error
	var tlsConfig *tls.Config

	secrets, err := secretProvider.GetSecret(config.Database.Type)
	if err != nil {
		return fmt.Errorf("unable to get Database Credentials for Store and Forward: %s", err.Error())
	}

	credentials := bootstrapConfig.Credentials{
		Username: secrets[secret.UsernameKey],
		Password: secrets[secret.PasswordKey],
	}

	if secret.IsSecurityEnabled() {
		tlsConfig, err = createRedisTlsConfig(v2Common.RedisTlsSecretName, secretProvider)
		if err != nil {
			logger.Errorf("couldn't create database tls cert and key: %v", err.Error())
			return err
		}
		logger.Info("create database tls configs successfully")
	}

	startup := environment.GetStartupInfo(svc.serviceKey)

	tryUntil := time.Now().Add(time.Duration(startup.Duration) * time.Second)

	var storeClient interfaces.StoreClient
	for time.Now().Before(tryUntil) {
		if storeClient, err = svc.createStoreClient(config.Database, credentials, tlsConfig); err != nil {
			logger.Warnf("unable to initialize Database '%s' for Store and Forward: %s", config.Database.Type, err.Error())
			time.Sleep(time.Duration(startup.Interval) * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return fmt.Errorf("initialize Database for Store and Forward failed: %s", err.Error())
	}

	svc.dic.Update(di.ServiceConstructorMap{
		container.StoreClientName: func(get di.Get) interface{} {
			return storeClient
		},
	})
	return nil
}

// createRedisTlsConfig get TLS certificates from secret provider and creates Redis TLS config
func createRedisTlsConfig(secretName string, provider bootstrapInterfaces.SecretProvider) (*tls.Config, errors.EdgeX) {
	var tlsConfig *tls.Config
	secrets, err := provider.GetSecret(secretName)
	if err != nil {
		return tlsConfig, errors.NewCommonEdgeX(errors.KindServerError, "fail to get the Redis TLS secrets from the secret provider", err)
	}

	// check if secretClientCert, secretClientKey and SecretCACert exist in the secrets
	if _, ok := secrets[messaging.SecretClientCert]; !ok {
		return tlsConfig, errors.NewCommonEdgeX(errors.KindServerError, "redis TLS client cert not found in the secret", err)
	}
	if _, ok := secrets[messaging.SecretClientKey]; !ok {
		return tlsConfig, errors.NewCommonEdgeX(errors.KindServerError, "redis TLS client key not found in the secret", err)
	}
	if _, ok := secrets[messaging.SecretCACert]; !ok {
		return tlsConfig, errors.NewCommonEdgeX(errors.KindServerError, "redis TLS client ca not found in the secret", err)
	}

	tlsConfig, err = v2Common.CreateRedisTlsConfigFromPEM([]byte(secrets[messaging.AuthModeCert]), []byte(secrets[messaging.SecretClientKey]), []byte(secrets[messaging.SecretCACert]))
	if err != nil {
		return tlsConfig, errors.NewCommonEdgeXWrapper(err)
	}
	return tlsConfig, nil
}
