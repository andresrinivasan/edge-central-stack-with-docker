// Copyright (C) 2021-2022 IOTech Ltd

package xpert

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/handler"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	// Name of the keys to look for in secret provider
	AWSIoTMQTTSecretClientCert = messaging.SecretClientCert
	AWSIoTMQTTSecretClientKey  = messaging.SecretClientKey
)

type AWSIoTCoreMQTTConfig struct {
	BaseMqttConfig
	// The name of the path in secret provider to retrieve your secrets, only take effect when SecretFormat=PEMBlock
	SecretPath string
	// SkipCertVerify
	SkipCertVerify bool
}

type AWSIoTCoreMQTTSender struct {
	BaseMqttSender
	awsIoTCoreConfig AWSIoTCoreMQTTConfig
}

type iotCoreSecrets struct {
	keyPEMBlock  []byte
	certPEMBlock []byte
}

func NewAWSIoTCoreMQTTSender(awsIoTCoreConfig AWSIoTCoreMQTTConfig, persistOnError, usingSharedClient bool) *AWSIoTCoreMQTTSender {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(awsIoTCoreConfig.BrokerAddress)
	opts.SetClientID(awsIoTCoreConfig.ClientId)
	opts.SetAutoReconnect(awsIoTCoreConfig.AutoReconnect)

	sender := &AWSIoTCoreMQTTSender{
		awsIoTCoreConfig: awsIoTCoreConfig,
		BaseMqttSender: BaseMqttSender{
			client:            nil,
			config:            awsIoTCoreConfig.BaseMqttConfig,
			persistOnError:    persistOnError,
			mqttClientFactory: NewAWSIoTCoreMqttClientFactory(awsIoTCoreConfig, opts),
			usingSharedClient: usingSharedClient,
		},
	}

	return sender
}

type AWSIoTCoreMqttClientFactory struct {
	opts           *MQTT.ClientOptions
	secretPath     string
	skipCertVerify bool
}

func NewAWSIoTCoreMqttClientFactory(config AWSIoTCoreMQTTConfig, opts *MQTT.ClientOptions) MqttClientFactory {
	return &AWSIoTCoreMqttClientFactory{
		opts:           opts,
		secretPath:     config.SecretPath,
		skipCertVerify: config.SkipCertVerify,
	}
}

func (factory *AWSIoTCoreMqttClientFactory) Create(ctx interfaces.AppFunctionContext) (MQTT.Client, error) {
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

func (factory *AWSIoTCoreMqttClientFactory) getSecrets(ctx interfaces.AppFunctionContext) (*iotCoreSecrets, error) {
	iotCoreSecrets := &iotCoreSecrets{}

	secrets, err := ctx.GetSecret(factory.secretPath)
	if err != nil {
		ctx.LoggingClient().Errorf("failed to get secrets from the secret provider. Error: %s", err)
		return nil, err
	}
	iotCoreSecrets.keyPEMBlock = []byte(secrets[AWSIoTMQTTSecretClientKey])
	iotCoreSecrets.certPEMBlock = []byte(secrets[AWSIoTMQTTSecretClientCert])

	return iotCoreSecrets, nil
}

func (factory *AWSIoTCoreMqttClientFactory) validateSecrets(secrets iotCoreSecrets) error {
	// need cert pem block to make a successful connection
	if len(secrets.certPEMBlock) <= 0 {
		return errors.New("cert PEM block is empty")
	}
	// need key pem block to make a successful connection
	if len(secrets.keyPEMBlock) <= 0 {
		return errors.New("key PEM block is empty")
	}
	return nil
}

func (factory *AWSIoTCoreMqttClientFactory) configureMQTTClientForAuth(secrets iotCoreSecrets) error {
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
	}
	factory.opts.SetTLSConfig(tlsConfig)

	return nil
}

func (sender *AWSIoTCoreMQTTSender) Send(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("no Event Received")
	}

	// When data comes from Batch function, its data type would be [][]byte.
	// When data comes from TransformToAWSDeviceShadow func or Store-Forward process, it must have been marshaled to []byte.
dataTypeDetection:
	switch d := data.(type) {
	case []byte:
		// determine the actual type of input data
		var batchedData []interface{}
		if err := json.Unmarshal(d, &batchedData); err == nil {
			data = batchedData
			goto dataTypeDetection
		} else {
			var singleData map[string]interface{}
			if err := json.Unmarshal(d, &singleData); err == nil {
				data = singleData
				goto dataTypeDetection
			} else {
				return false, errors.New("unsupported data type")
			}
		}
	case []interface{}:
		// export batched data. Per Device Shadow's property, batched data have to be published one by one.
		var dataToBeResent [][]byte
		for _, data := range d {
			byteData, err := util.CoerceType(data)
			if err != nil {
				return false, fmt.Errorf("unsupported data type passed in: %v. Error: %s",
					reflect.TypeOf(data), err)
			}
			if err := sender.mqttPublish(byteData, ctx); err != nil {
				dataToBeResent = append(dataToBeResent, byteData)
			}
		}
		if len(dataToBeResent) > 0 {
			retryData, err := json.Marshal(dataToBeResent)
			if err != nil {
				return false, fmt.Errorf("failed to marshal the retryData, error: %e", err)
			}
			sender.setRetryData(ctx, retryData)
			return false, errors.New("persisting data for later retry")
		}
	default:
		// export single data
		exportData, err := json.Marshal(data)
		if err != nil {
			ctx.LoggingClient().Errorf("marshaling input data to JSON failed, passed in data "+
				"must support marshaling to JSON", err)
			return false, err
		}
		if err = sender.mqttPublish(exportData, ctx); err != nil {
			ctx.LoggingClient().Errorf("failed to publish data to MQTT broker %s. Error: %s",
				sender.awsIoTCoreConfig.BrokerAddress, err)
			sender.setRetryData(ctx, exportData)
			return false, errors.New("persisting data for later retry")
		}
	}

	ctx.LoggingClient().Debug("Sent data to AWS IoT Core MQTT Broker.")
	ctx.LoggingClient().Trace("Data exported", "Transport", "MQTT", common.CorrelationHeader,
		ctx.CorrelationID)
	return true, nil
}

func (sender *AWSIoTCoreMQTTSender) mqttPublish(exportData []byte, ctx interfaces.AppFunctionContext) error {
	// check if client haven't be initialized OR the cache has been invalidated (due to new/updated secrets), need to (re)initialize the client
	if sender.client == nil || sender.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		err := sender.initializeMQTTClient(ctx)
		if err != nil {
			return err
		}
	}
	if err := sender.publish(ctx, exportData); err != nil {
		return err
	}

	return nil
}

type AWSIoTCoreResponse struct {
	Status       int
	ResponseBody interface{}
}

// SendResponse sends the result of core command execution to AWS IoT Core.
// This function should only be invoked right after the function ExecuteCoreCommand, which will send the error in the type of EdgeX Error.
func (sender *AWSIoTCoreMQTTSender) SendResponse(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("no Event Received")
	}

	res := AWSIoTCoreResponse{}
	switch data := data.(type) {
	case edgexErrors.EdgeX:
		res.ResponseBody = data.Error()
		res.Status = data.Code()
	default:
		res.ResponseBody = data
		res.Status = http.StatusOK
	}
	return sender.Send(ctx, res)
}
