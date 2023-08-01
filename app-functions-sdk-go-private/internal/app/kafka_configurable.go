// Copyright (C) 2021 IOTech Ltd

package app

import (
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
)

// Define configurable parameters for kafka send function
const (
	KafkaClientID  = "clientid"
	KafkaAddress   = "address"
	KafkaPort      = "port"
	KafkaTopic     = "topic"
	KafkaPartition = "partition"
)

func (app *Configurable) KafkaSend(parameters map[string]string) interfaces.AppFunction {
	clientID, ok := parameters[KafkaClientID]
	if !ok {
		app.lc.Error("mandatory parameter not found:" + KafkaClientID)
		return nil
	}
	address, ok := parameters[KafkaAddress]
	if !ok {
		app.lc.Error("mandatory parameter not found:" + KafkaAddress)
		return nil
	}
	portStr, ok := parameters[KafkaPort]
	if !ok {
		app.lc.Error("mandatory parameter not found:" + KafkaPort)
		return nil
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		app.lc.Error("failed to parse Port:" + portStr)
		return nil
	}
	topic, ok := parameters[KafkaTopic]
	if !ok {
		app.lc.Error("mandatory parameter not found:" + KafkaTopic)
		return nil
	}

	//if Partition is not set, use partition 0 as default
	partition := 0
	partitionStr, ok := parameters[KafkaPartition]
	if ok {
		// following usage of strconv.Atoi is complained by lint for
		// 'G109: Potential Integer overflow made by strconv.Atoi result conversion to int16/32 (gosec)'.
		// Note that partition is declared as int, which exactly matches to the return type of strconv.Atoi, so I don't
		// think that there is any potential int overflow in this case.  Mark nolint to ignore the warning.
		partition, err = strconv.Atoi(partitionStr) //nolint: gosec
		if err != nil {
			app.lc.Errorf("could not parse '%s' to a int for '%s' parameter.  Error:%v", partitionStr, KafkaPartition, err)
			return nil
		}
	} else {
		app.lc.Info("Partition is not set, use partition 0 as default.")
	}

	// PersistOnError is optional and is false by default.
	// If the Kafka send fails and persistOnError is true and StoreAndForward is enabled,
	// the data will be stored for later retry.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		var err error
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			app.lc.Errorf("could not parse '%s' to a bool for '%s' parameter.  Error: %v", value, PersistOnError, err)
			return nil
		}
	}

	config := xpert.KafkaEndpoint{
		ClientID:  clientID,
		Address:   address,
		Port:      port,
		Topic:     topic,
		Partition: int32(partition), //nolint: gosec
	}

	// skipCertVerify is optional and is false by default.
	skipCertVerify := false
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			app.lc.Errorf("could not parse '%s' to a bool for '%s' parameter.  Error: %v", skipVerifyVal, SkipVerify, err)
			return nil
		}
	} else {
		app.lc.Infof("%s is not set, use %v as default.", SkipVerify, skipCertVerify)
	}

	// authMode is optional and is none by default.
	authMode, ok := parameters[AuthMode]
	if ok {
		authMode = strings.ToLower(authMode)
		// Check if supported authMode is specified; KafkaSend only supports two-way TLS client authentication(clientcert) as of release 1.7.1
		if authMode != messaging.AuthModeCert && authMode != messaging.AuthModeNone {
			app.lc.Errorf("parameter %s specifies unsupported value %s.", AuthMode, authMode)
			return nil
		}
	} else {
		app.lc.Debugf("%s parameter is not set, use %s as default", AuthMode, messaging.AuthModeNone)
		authMode = messaging.AuthModeNone
	}

	secretsConfig := xpert.KafkaSecretsConfig{
		AuthMode:       authMode,
		SkipCertVerify: skipCertVerify,
	}

	if authMode == messaging.AuthModeCert {
		secretPath, ok := parameters[SecretPath]
		if !ok {
			app.lc.Errorf("mandatory parameter %s not found.", SecretPath)
			return nil
		}
		secretsConfig.SecretPath = secretPath
	}

	sender := xpert.NewKafkaSender(config, secretsConfig, persistOnError)
	return sender.KafkaSend
}
