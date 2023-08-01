// Copyright (C) 2021-2022 IOTech Ltd

package xpert

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/handler"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	// Name of the keys to look for in secret provider
	IoTHubSecretClientKey  = messaging.SecretClientKey
	IoTHubSecretClientCert = messaging.SecretClientCert
	IoTHubSecretUsername   = messaging.SecretUsernameKey

	ridPrefix                  = "?$rid="
	directMethodsResponseTopic = "$iothub/methods/res/" + MQTTTopicPlaceholderStatus + "/" + ridPrefix +
		MQTTTopicPlaceholderRequestId
)

type AzureIoTHubMQTTConfig struct {
	// Common Mqtt Config as embedded field
	BaseMqttConfig
	// The name of the path in secret provider to retrieve your secrets, only take effect when IotHubSecretFormat=PEMBlock
	SecretPath string
	// AuthMode indicates what authentication mode to connect to Azure IOT Hub when running application service.
	// As of release of Edge Xpert 1.7.0, only clientcert is supported.
	AuthMode string
	// SkipCertVerify
	SkipCertVerify bool
}

type AzureIoTHubMQTTSender struct {
	BaseMqttSender
	iotHubConfig AzureIoTHubMQTTConfig
}

type iotHubSecrets struct {
	keyPEMBlock  []byte
	certPEMBlock []byte
	username     string
}

// NewAzureIoTHubMQTTSender - factory method for creating azureIotHubSender
func NewAzureIoTHubMQTTSender(iotHubConfig AzureIoTHubMQTTConfig, persistOnError bool, usingSharedClient bool) *AzureIoTHubMQTTSender {

	opts := MQTT.NewClientOptions()
	opts.AddBroker(iotHubConfig.BrokerAddress)
	opts.SetClientID(iotHubConfig.ClientId)
	opts.SetProtocolVersion(4) // 4 = MQTT 3.1.1, which is the broker version of Azure IOT Hub
	opts.SetAutoReconnect(iotHubConfig.AutoReconnect)

	//avoid casing issues
	iotHubConfig.AuthMode = strings.ToLower(iotHubConfig.AuthMode)

	sender := &AzureIoTHubMQTTSender{
		iotHubConfig: iotHubConfig,
		BaseMqttSender: BaseMqttSender{
			client:            nil,
			config:            iotHubConfig.BaseMqttConfig,
			persistOnError:    persistOnError,
			mqttClientFactory: NewAzureIoTHubMqttClientFactory(iotHubConfig, opts),
			usingSharedClient: usingSharedClient,
		},
	}

	return sender
}

type IoTHubMqttClientFactory struct {
	opts           *MQTT.ClientOptions
	authMode       string
	secretPath     string
	skipCertVerify bool
}

func NewAzureIoTHubMqttClientFactory(config AzureIoTHubMQTTConfig, opts *MQTT.ClientOptions) MqttClientFactory {
	return &IoTHubMqttClientFactory{
		opts:           opts,
		authMode:       config.AuthMode,
		secretPath:     config.SecretPath,
		skipCertVerify: config.SkipCertVerify,
	}
}

func (factory *IoTHubMqttClientFactory) Create(ctx interfaces.AppFunctionContext) (MQTT.Client, error) {
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

func (factory *IoTHubMqttClientFactory) getSecrets(ctx interfaces.AppFunctionContext) (*iotHubSecrets, error) {
	var keyPEMBlock []byte
	var certPEMBlock []byte
	secrets, err := ctx.GetSecret(factory.secretPath)
	if err != nil {
		return nil, err
	}
	keyPEMBlock = []byte(secrets[IoTHubSecretClientKey])
	certPEMBlock = []byte(secrets[IoTHubSecretClientCert])
	iotHubSecrets := &iotHubSecrets{
		keyPEMBlock:  keyPEMBlock,
		certPEMBlock: certPEMBlock,
		username:     secrets[IoTHubSecretUsername],
	}
	return iotHubSecrets, nil
}

func (factory *IoTHubMqttClientFactory) validateSecrets(secrets iotHubSecrets) error {
	if len(secrets.keyPEMBlock) <= 0 || len(secrets.certPEMBlock) <= 0 {
		return fmt.Errorf("AuthMode:%s selected however the key or cert PEM block was not found at secret path", IoTHubSecretClientCert)
	}
	if len(secrets.username) == 0 {
		return errors.New("username was not found at secret path")
	}
	return nil
}

func (factory *IoTHubMqttClientFactory) configureMQTTClientForAuth(secrets iotHubSecrets) error {
	// validate secrets prior to configure TLS
	err := factory.validateSecrets(secrets)
	if err != nil {
		return err
	}

	var cert tls.Certificate
	cert, err = tls.X509KeyPair(secrets.certPEMBlock, secrets.keyPEMBlock)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		ClientCAs:          nil,
		InsecureSkipVerify: factory.skipCertVerify, //nolint: gosec
		Certificates:       []tls.Certificate{cert},
		Renegotiation:      tls.RenegotiateOnceAsClient, // this must be specified for Azure IOT Hub
	}
	factory.opts.SetTLSConfig(tlsConfig)
	factory.opts.SetUsername(secrets.username)

	return nil
}

func (sender *AzureIoTHubMQTTSender) Send(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("no event received")
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

// SendDirectMethodResponse sends the result of core command execution to Azure IoT Hub as the response of Direct Method invocation.
// This function should only be invoked right after the function ExecuteCoreCommand, which will send the error in the type of EdgeX Error.
func (sender *AzureIoTHubMQTTSender) SendDirectMethodResponse(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("no data Received")
	}

	var status int
	var resBody interface{}
	if edgexErr, ok := data.(edgexErrors.EdgeX); ok {
		// Azure IoT Hub requires that the payload must be a document of type JSON
		resBody = fmt.Sprintf(`{"error":"%s"}`, edgexErr.Error())
		status = edgexErr.Code()
	} else {
		resBody = data
		status = http.StatusOK
	}

	receivedTopic, ok := ctx.GetValue(interfaces.RECEIVEDTOPIC)
	if !ok {
		return false, errors.New("received topic was not found in AppFunctionContext")
	}
	requestId := receivedTopic[strings.LastIndex(receivedTopic, ridPrefix)+len(ridPrefix):]

	// Refer to https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-devguide-direct-methods#response-1, the response
	// should be sent to $iothub/methods/res/{status}/?$rid={requestid}
	// Note that the value of {status} only accepts an integer, and the value of {requestid} is from the method invocation received from IoT Hub.
	topic := directMethodsResponseTopic
	topic = strings.Replace(topic, MQTTTopicPlaceholderStatus, strconv.Itoa(status), -1)
	topic = strings.Replace(topic, MQTTTopicPlaceholderRequestId, requestId, -1)
	sender.config.Topic = topic

	return sender.Send(ctx, resBody)
}
