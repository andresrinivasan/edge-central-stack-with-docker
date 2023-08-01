// Copyright (C) 2021-2022 IOTech Ltd

package xpert

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/handler"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// Deprecated: IBM Watson IoT Platform is being retired on December 1st, 2023. https://lp2.email.ibm.com/index.php/email/emailWebview?md_id=90149
type IBMWatsonMQTTConfig struct {
	BaseMqttConfig
	// The name of the path in secret provider to retrieve your secrets, only take effect when WatsonCertFormat=PEMBlock
	SecretPath string
	// SkipCertVerify
	SkipCertVerify bool
}

// Deprecated: IBM Watson IoT Platform is being retired on December 1st, 2023. https://lp2.email.ibm.com/index.php/email/emailWebview?md_id=90149
type IBMWatsonMQTTSender struct {
	BaseMqttSender
	watsonMqttConfig IBMWatsonMQTTConfig
}

type watsonSecrets struct {
	username     string
	password     string
	certpemblock []byte
}

// NewIBMWatsonMQTTSender - factory method for creation of IBMWatsonMQTTSender
// Deprecated: IBM Watson IoT Platform is being retired on December 1st, 2023. https://lp2.email.ibm.com/index.php/email/emailWebview?md_id=90149
func NewIBMWatsonMQTTSender(mqttConfig IBMWatsonMQTTConfig, persistOnError bool, usingSharedClient bool) *IBMWatsonMQTTSender {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttConfig.BrokerAddress)
	opts.SetClientID(mqttConfig.ClientId)
	opts.SetAutoReconnect(mqttConfig.AutoReconnect)

	//avoid casing issues
	sender := &IBMWatsonMQTTSender{
		watsonMqttConfig: mqttConfig,
		BaseMqttSender: BaseMqttSender{
			client:            nil,
			config:            mqttConfig.BaseMqttConfig,
			persistOnError:    persistOnError,
			mqttClientFactory: NewIBMWatsonMqttClientFactory(mqttConfig, opts),
			usingSharedClient: usingSharedClient,
		},
	}

	return sender
}

type IBMWatsonMqttClientFactory struct {
	opts           *MQTT.ClientOptions
	secretPath     string
	skipCertVerify bool
}

// Deprecated: IBM Watson IoT Platform is being retired on December 1st, 2023. https://lp2.email.ibm.com/index.php/email/emailWebview?md_id=90149
func NewIBMWatsonMqttClientFactory(config IBMWatsonMQTTConfig, opts *MQTT.ClientOptions) MqttClientFactory {
	return &IBMWatsonMqttClientFactory{
		opts:           opts,
		secretPath:     config.SecretPath,
		skipCertVerify: config.SkipCertVerify,
	}
}

func (factory *IBMWatsonMqttClientFactory) Create(ctx interfaces.AppFunctionContext) (MQTT.Client, error) {
	//get the secrets from the secret provider and populate the struct
	secrets, err := factory.getSecrets(ctx)
	if err != nil {
		return nil, err
	}
	//ensure that the required secret values exist
	if secrets != nil {
		// configure the mqtt client with the retrieved secret values
		err = factory.configureMQTTClientForAuth(*secrets)
		if err != nil {
			return nil, err
		}
	}

	factory.opts.OnConnectionLost = handler.MqttConnectionLostHandler(ctx.LoggingClient(), ctx.NotificationClient(),
		ctx.GetAllValues()[strings.ToLower(handler.ServiceKey)],
		ctx.GetAllValues()[strings.ToLower(handler.PostDisconnectionAlert)])

	return MQTT.NewClient(factory.opts), nil
}

func (factory *IBMWatsonMqttClientFactory) getSecrets(ctx interfaces.AppFunctionContext) (*watsonSecrets, error) {
	secrets, err := ctx.GetSecret(factory.secretPath)
	if err != nil {
		return nil, err
	}
	mqttSecrets := &watsonSecrets{
		username: secrets[MQTTSecretUsername],
		password: secrets[MQTTSecretPassword],
	}
	mqttSecrets.certpemblock = []byte(secrets[MQTTSecretCACert])
	return mqttSecrets, nil
}

func (factory *IBMWatsonMqttClientFactory) validateSecrets(secrets watsonSecrets) error {
	// need username to make a successful connection
	if len(secrets.username) <= 0 {
		return errors.New("mandatory username is empty")
	}
	// need password to make a successful connection
	if len(secrets.password) <= 0 {
		return errors.New("mandatory password is empty")
	}
	return nil
}

func (factory *IBMWatsonMqttClientFactory) configureMQTTClientForAuth(secrets watsonSecrets) error {
	// validate secrets prior to configure TLS
	err := factory.validateSecrets(secrets)
	if err != nil {
		return err
	}

	caCertPool := x509.NewCertPool()
	if len(secrets.certpemblock) > 0 {
		if ok := caCertPool.AppendCertsFromPEM(secrets.certpemblock); !ok {
			return errors.New("error parsing CA certificate PEM block")
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: factory.skipCertVerify, //nolint: gosec
		ClientCAs:          caCertPool,
	}
	factory.opts.SetTLSConfig(tlsConfig)
	factory.opts.SetUsername(secrets.username)
	factory.opts.SetPassword(secrets.password)

	return nil
}

// Send sends data from the previous function to the specified IBM Watson MQTT broker.
// If no previous function exists, then the event that triggered the pipeline will be used.
func (sender *IBMWatsonMQTTSender) Send(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("no Event Received")
	}

	exportData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}

	// check if client haven't be initialized OR the cache has been invalidated (due to new/updated secrets), need to (re)initialize the client
	if sender.client == nil || sender.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		err := sender.initializeMQTTClient(ctx)
		if err != nil {
			return false, err
		}
	}
	if err = sender.publish(ctx, exportData); err != nil {
		return false, err
	}

	return true, nil
}
