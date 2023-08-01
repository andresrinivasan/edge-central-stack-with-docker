// Copyright (C) 2021-2022 IOTech Ltd

package xpert

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/handler"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type RegularMQTTConfig struct {
	BaseMqttConfig
	// The name of the path in secret provider to retrieve secrets
	SecretPath string
	// SkipCertVerify indicates whether to verify the server's certificate chain and host name
	SkipCertVerify bool
	// AuthMode indicates what to use when connecting to the broker. Options are "none", "cacert" , "usernamepassword", "clientcert".
	// If a CA Cert exists in the SecretPath then it will be used for all modes except "none".
	AuthMode string
}

// RegularMQTTSender ...
type RegularMQTTSender struct {
	BaseMqttSender
}

// NewRegularMQTTSender - factory method for creation of RegularMQTTSender
func NewRegularMQTTSender(mqttConfig RegularMQTTConfig, persistOnError bool, usingSharedClient bool) *RegularMQTTSender {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttConfig.BrokerAddress)
	opts.SetClientID(mqttConfig.ClientId)
	opts.SetAutoReconnect(mqttConfig.AutoReconnect)

	sender := &RegularMQTTSender{
		BaseMqttSender: BaseMqttSender{
			client:            nil,
			config:            mqttConfig.BaseMqttConfig,
			persistOnError:    persistOnError,
			mqttClientFactory: NewRegularMqttClientFactory(mqttConfig, opts),
			usingSharedClient: usingSharedClient,
		},
	}

	return sender
}

type RegularMqttClientFactory struct {
	opts           *MQTT.ClientOptions
	authMode       string
	secretPath     string
	skipCertVerify bool
}

func NewRegularMqttClientFactory(config RegularMQTTConfig, opts *MQTT.ClientOptions) MqttClientFactory {
	return &RegularMqttClientFactory{
		opts:           opts,
		secretPath:     config.SecretPath,
		skipCertVerify: config.SkipCertVerify,
		authMode:       config.AuthMode,
	}
}

func (factory *RegularMqttClientFactory) Create(ctx interfaces.AppFunctionContext) (MQTT.Client, error) {
	if factory.authMode == "" {
		factory.authMode = messaging.AuthModeNone
		ctx.LoggingClient().Warnf("AuthMode not set, defaulting to %s", messaging.AuthModeNone)
	}

	// get the secrets from the secret provider and populate the struct
	secrets, err := factory.getSecrets(ctx)
	if err != nil {
		return nil, err
	}
	// ensure that the required secret values exist
	if secrets != nil {
		err = factory.validateSecrets(*secrets)
		if err != nil {
			return nil, err
		}
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

type mqttSecrets struct {
	username     string
	password     string
	keyPemBlock  []byte
	certPemBlock []byte
	caPemBlock   []byte
}

func (factory *RegularMqttClientFactory) getSecrets(ctx interfaces.AppFunctionContext) (*mqttSecrets, error) {
	// No Auth? No Problem!...No secrets required.
	if factory.authMode == messaging.AuthModeNone {
		return nil, nil
	}

	secrets, err := ctx.GetSecret(factory.secretPath)
	if err != nil {
		return nil, err
	}
	mqttSecrets := &mqttSecrets{
		username:     secrets[messaging.SecretUsernameKey],
		password:     secrets[messaging.SecretPasswordKey],
		keyPemBlock:  []byte(secrets[messaging.SecretClientKey]),
		certPemBlock: []byte(secrets[messaging.SecretClientCert]),
		caPemBlock:   []byte(secrets[messaging.SecretCACert]),
	}

	return mqttSecrets, nil
}

func (factory RegularMqttClientFactory) validateSecrets(secrets mqttSecrets) error {
	caCertPool := x509.NewCertPool()
	switch factory.authMode {
	case messaging.AuthModeUsernamePassword:
		if secrets.username == "" || secrets.password == "" {
			return fmt.Errorf("auth mode %s selected however Username or Password was not found "+
				"at secret path", factory.authMode)
		}
	case messaging.AuthModeCert:
		// need both to make a successful connection
		if len(secrets.keyPemBlock) <= 0 || len(secrets.certPemBlock) <= 0 {
			return fmt.Errorf("auth mode %s selected however the key or cert PEM block "+
				"was not found at secret path", factory.authMode)
		}
	case messaging.AuthModeCA:
		if len(secrets.caPemBlock) <= 0 {
			return fmt.Errorf("auth mode %s selected however no PEM Block was found "+
				"at secret path", factory.authMode)
		}
	case messaging.AuthModeNone:
		return errors.New("invalid AuthMode selected")
	default:
		return fmt.Errorf("auth mode %s is not supported", factory.authMode)
	}

	if len(secrets.caPemBlock) > 0 {
		ok := caCertPool.AppendCertsFromPEM(secrets.caPemBlock)
		if !ok {
			return errors.New("error parsing CA Certificate")
		}
	}

	return nil
}

func (factory RegularMqttClientFactory) configureMQTTClientForAuth(secrets mqttSecrets) error {
	var cert tls.Certificate
	var err error
	caCertPool := x509.NewCertPool()
	tlsConfig := &tls.Config{
		InsecureSkipVerify: factory.skipCertVerify, //nolint: gosec
	}
	// Username may be required for cert authentication
	if secrets.username != "" {
		factory.opts.SetUsername(secrets.username)
	}
	switch factory.authMode {
	case messaging.AuthModeUsernamePassword:
		factory.opts.SetPassword(secrets.password)
	case messaging.AuthModeCert:
		cert, err = tls.X509KeyPair(secrets.certPemBlock, secrets.keyPemBlock)
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	case messaging.AuthModeCA:
		break
	case messaging.AuthModeNone:
		return nil
	default:
		return fmt.Errorf("auth mode %s is not supported", factory.authMode)
	}

	if len(secrets.caPemBlock) > 0 {
		ok := caCertPool.AppendCertsFromPEM(secrets.caPemBlock)
		if !ok {
			return errors.New("error parsing CA PEM block")
		}
		tlsConfig.RootCAs = caCertPool
	}

	factory.opts.SetTLSConfig(tlsConfig)

	return nil
}

// Send sends data from the previous function to the specified MQTT broker.
// If no previous function exists, then the event that triggered the pipeline will be used.
// If there is a pipeline using MQTT trigger, this function will use the same MQTT client initialized by the trigger.
func (sender *RegularMQTTSender) Send(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
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
