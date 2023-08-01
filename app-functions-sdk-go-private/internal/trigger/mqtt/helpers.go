// Copyright (C) 2021 IOTech Ltd

package mqtt

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/secure"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
)

// Due to the import cycle issue, the LoadAWSMQTTConfig function in the app package cannot be reused in this package.
// As a workaround, use newAWSIoTCoreMQTTConfig helper function to convert ExternalMqttConfig to AWSIoTCoreMQTTConfig.
// In contrast to LoadAWSMQTTConfig, the newAWSIoTCoreMQTTConfig function doesn't check configuration parameters.
// However, if the configuration parameters goes wrong the MQTT client creation process will return an error to stop the trigger initialization.
func newAWSIoTCoreMQTTConfig(config sdkCommon.ExternalMqttConfig) xpert.AWSIoTCoreMQTTConfig {
	return xpert.AWSIoTCoreMQTTConfig{
		SecretPath:     config.SecretPath,
		SkipCertVerify: config.SkipCertVerify,
		BaseMqttConfig: xpert.BaseMqttConfig{
			BrokerAddress: config.Url,
			ClientId:      config.ClientId,
			AutoReconnect: config.AutoReconnect,
			QoS:           config.QoS,
			Retain:        config.Retain,
		},
	}
}

// Due to the import cycle issue, the LoadAzureMQTTConfig function in the app package cannot be reused in this package.
// As a workaround, use newAzureIoTHubMQTTConfig helper function to convert ExternalMqttConfig to AzureIoTHubMQTTConfig.
// In contrast to LoadAzureMQTTConfig, the newAzureIoTHubMQTTConfig function doesn't check configuration parameters.
// However, if the configuration parameters goes wrong the MQTT client creation process will return an error to stop the trigger initialization.
func newAzureIoTHubMQTTConfig(config sdkCommon.ExternalMqttConfig) xpert.AzureIoTHubMQTTConfig {
	return xpert.AzureIoTHubMQTTConfig{
		SecretPath:     config.SecretPath,
		SkipCertVerify: config.SkipCertVerify,
		AuthMode:       config.AuthMode,
		BaseMqttConfig: xpert.BaseMqttConfig{
			BrokerAddress: config.Url,
			ClientId:      config.ClientId,
			AutoReconnect: config.AutoReconnect,
			QoS:           config.QoS,
			Retain:        config.Retain,
		},
	}
}

func createMQTTClient(lc logger.LoggingClient, dic *di.Container, opts *pahoMqtt.ClientOptions) (pahoMqtt.Client, error) {
	config := container.ConfigurationFrom(dic.Get)
	brokerConfig := config.Trigger.ExternalMqtt

	// Since this factory is shared between the MQTT pipeline function and this trigger we must provide
	// a dummy AppFunctionContext which will provide access to GetSecret
	appContext := appfunction.NewContext("", dic, "")
	var mqttClient pahoMqtt.Client
	var err error
	if strings.Contains(brokerConfig.Url, xpert.DomainAWS) {
		awsMQTTConfig := newAWSIoTCoreMQTTConfig(brokerConfig)
		mqttClient, err = xpert.NewAWSIoTCoreMqttClientFactory(awsMQTTConfig, opts).Create(appContext)
	} else if strings.Contains(brokerConfig.Url, xpert.DomainAzure) {
		azureMQTTConfig := newAzureIoTHubMQTTConfig(brokerConfig)
		mqttClient, err = xpert.NewAzureIoTHubMqttClientFactory(azureMQTTConfig, opts).Create(appContext)
	} else {
		mqttFactory := secure.NewMqttFactory(
			appContext,
			lc,
			brokerConfig.AuthMode,
			brokerConfig.SecretPath,
			brokerConfig.SkipCertVerify,
		)
		mqttClient, err = mqttFactory.Create(opts)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create MQTT Client: %s", err.Error())
	}

	lc.Infof("Connecting to mqtt broker for MQTT trigger at: %s", brokerConfig.Url)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("could not connect to broker for MQTT trigger: %s", token.Error().Error())
	}

	lc.Info("Connected to mqtt server for MQTT trigger")

	return mqttClient, nil
}
