// Copyright (C) 2021-2023 IOTech Ltd

package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func LoadRegularMQTTConfig(loggingClient logger.LoggingClient, parameters map[string]string) (xpert.RegularMQTTConfig, bool, error) {
	config := xpert.RegularMQTTConfig{}

	// PersistOnError is optional and is false by default.
	// If the send fails to send and persistOnError is true and StoreAndForward is enabled,
	// the data will be stored for later retry.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		var err error
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			return config, persistOnError, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter. "+
				"error: %s", value, PersistOnError, err)
		}
	}

	brokerAddress, ok := parameters[BrokerAddress]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", BrokerAddress)
	}

	topic, ok := parameters[Topic]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", BrokerAddress)
	}

	var err error
	qos := 0
	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", Qos, err)
		}
		if qos < 0 || qos > 2 {
			return config, persistOnError, fmt.Errorf("invalid %s value: %d", Qos, qos)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", Qos, qos)
	}

	autoReconnect := false
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s",
				AutoReconnect, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", AutoReconnect, autoReconnect)
	}

	retain := false
	retainVal, ok := parameters[Retain]
	if ok {
		retain, err = strconv.ParseBool(retainVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", Retain, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", Retain, retain)
	}

	skipCertVerify := false
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", SkipVerify, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", SkipVerify, skipCertVerify)
	}

	authMode, ok := parameters[AuthMode]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", AuthMode)
	}

	secretPath := ""
	if authMode != messaging.AuthModeNone {
		secretPath, ok = parameters[SecretPath]
		if !ok {
			return config, persistOnError, fmt.Errorf("authmod %s selected, but parameter %s not found",
				authMode, SecretPath)
		}
	}

	clientID := parameters[ClientID]

	config = xpert.RegularMQTTConfig{
		BaseMqttConfig: xpert.BaseMqttConfig{
			Retain:        retain,
			AutoReconnect: autoReconnect,
			QoS:           byte(qos),
			BrokerAddress: brokerAddress,
			ClientId:      clientID,
			Topic:         topic,
		},
		SkipCertVerify: skipCertVerify,
		SecretPath:     secretPath,
		AuthMode:       authMode,
	}
	return config, persistOnError, nil
}

func LoadSharedMQTTConfig(mqttConfig common.ExternalMqttConfig, parameters map[string]string) map[string]string {
	parameters[BrokerAddress] = mqttConfig.Url
	parameters[ClientID] = mqttConfig.ClientId
	parameters[Qos] = strconv.Itoa(int(mqttConfig.QoS))
	parameters[Retain] = strconv.FormatBool(mqttConfig.Retain)
	parameters[AutoReconnect] = strconv.FormatBool(mqttConfig.AutoReconnect)
	parameters[SkipVerify] = strconv.FormatBool(mqttConfig.SkipCertVerify)
	parameters[AuthMode] = mqttConfig.AuthMode
	parameters[SecretPath] = mqttConfig.SecretPath
	return parameters
}

func LoadAWSMQTTConfig(loggingClient logger.LoggingClient, parameters map[string]string) (xpert.AWSIoTCoreMQTTConfig, bool, error) {
	config := xpert.AWSIoTCoreMQTTConfig{}

	// PersistOnError is optional and is false by default.
	// If the AWS send fails and persistOnError is true and StoreAndForward is enabled,
	// the data will be stored for later retry.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		var err error
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			if err != nil {
				return config, persistOnError, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter. "+
					"error: %s", value, PersistOnError, err)
			}
		}
	}

	brokerAddress, ok := parameters[BrokerAddress]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", BrokerAddress)
	}
	topic, ok := parameters[Topic]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", Topic)
	}
	clientID, ok := parameters[ClientID]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", clientID)
	}
	qos := 0
	var err error
	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", Qos, err)
		}
		// AWS IoT only supports QoS 0 and 1
		if qos > 1 {
			return config, persistOnError, fmt.Errorf("AWS IoT doesn't support QoS: %d", qos)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", Qos, qos)
	}

	autoReconnect := false
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s",
				AutoReconnect, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", AutoReconnect, autoReconnect)
	}

	skipCertVerify := false
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", SkipVerify, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", SkipVerify, skipCertVerify)
	}

	config = xpert.AWSIoTCoreMQTTConfig{
		SkipCertVerify: skipCertVerify,
		BaseMqttConfig: xpert.BaseMqttConfig{
			BrokerAddress: brokerAddress,
			ClientId:      clientID,
			Topic:         topic,
			AutoReconnect: autoReconnect,
			QoS:           byte(qos),
			//Per https://docs.aws.amazon.com/iot/latest/developerguide/mqtt.html, AWS IoT does not support
			//retained messages, so set retain to false always.
			Retain: false,
		},
	}

	secretPath, ok := parameters[SecretPath]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", SecretPath)
	}
	config.SecretPath = secretPath
	return config, persistOnError, nil
}

func LoadAzureMQTTConfig(loggingClient logger.LoggingClient, parameters map[string]string) (xpert.AzureIoTHubMQTTConfig, bool, error) {
	config := xpert.AzureIoTHubMQTTConfig{}

	// PersistOnError is optional and is false by default.
	// If Send fails and persistOnError is true and StoreAndForward is enabled,
	// the data will be stored for later retry.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		var err error
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			return config, persistOnError, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter. "+
				"error: %s", value, PersistOnError, err)
		}
	}

	brokerAddr, ok := parameters[BrokerAddress]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", BrokerAddress)
	}
	topic, ok := parameters[Topic]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", Topic)
	}
	clientID, ok := parameters[ClientID]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", ClientID)
	}
	var err error
	qos := 0
	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", Qos, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", Qos, qos)
	}

	retain := false
	retainVal, ok := parameters[Retain]
	if ok {
		retain, err = strconv.ParseBool(retainVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", Retain, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", Retain, retain)
	}

	autoReconnect := false
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", AutoReconnect, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", AutoReconnect, autoReconnect)
	}

	skipCertVerify := false
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", SkipVerify, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", SkipVerify, skipCertVerify)
	}

	config = xpert.AzureIoTHubMQTTConfig{
		SkipCertVerify: skipCertVerify,
		BaseMqttConfig: xpert.BaseMqttConfig{
			BrokerAddress: brokerAddr,
			ClientId:      clientID,
			Topic:         topic,
			AutoReconnect: autoReconnect,
			QoS:           byte(qos),
			Retain:        retain,
		},
	}

	secretPath, ok := parameters[SecretPath]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", SecretPath)
	}
	authMode, ok := parameters[AuthMode]
	if !ok {
		return config, persistOnError, fmt.Errorf("mandatory parameter %s not found", AuthMode)
	} else {
		authMode = strings.ToLower(authMode)
		// Check if supported authMode is specified
		if authMode != messaging.AuthModeCert {
			return config, persistOnError, fmt.Errorf("parameter %s specifies unsupported value %s",
				AuthMode, authMode)
		}
	}
	config.SecretPath = secretPath
	config.AuthMode = authMode

	return config, persistOnError, nil
}

func LoadIBMMQTTConfig(loggingClient logger.LoggingClient, parameters map[string]string) (xpert.IBMWatsonMQTTConfig, bool, error) { // nolint: staticcheck
	config := xpert.IBMWatsonMQTTConfig{} // nolint: staticcheck

	// PersistOnError is optional and is false by default.
	// If the Watson send fails and persistOnError is true and StoreAndForward is enabled,
	// the data will be stored for later retry.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		var err error
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			return config, persistOnError, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter", value, PersistOnError)
		}
	}

	brokerAddr, ok := parameters[BrokerAddress]
	if !ok {
		return config, false, fmt.Errorf("mandatory parameter %s not found", BrokerAddress)
	}
	topic, ok := parameters[Topic]
	if !ok {
		return config, false, fmt.Errorf("mandatory parameter %s not found", Topic)
	}
	clientID, ok := parameters[ClientID]
	if !ok {
		return config, false, fmt.Errorf("mandatory parameter %s not found", ClientID)
	}
	secretPath, ok := parameters[SecretPath]
	if !ok {
		return config, false, fmt.Errorf("mandatory parameter %s not found", SecretPath)
	}
	var err error
	qos := 0
	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", Qos, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", Qos, qos)
	}

	retain := false
	retainVal, ok := parameters[Retain]
	if ok {
		retain, err = strconv.ParseBool(retainVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", Retain, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", Retain, retain)
	}

	autoReconnect := false
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("unable to parse %s value. error: %s", AutoReconnect, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", AutoReconnect, autoReconnect)
	}

	skipCertVerify := false
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			return config, persistOnError, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter. "+
				"error: %s", skipVerifyVal, SkipVerify, err)
		}
	} else {
		loggingClient.Infof("%s is not set, use %v as default.", SkipVerify, skipCertVerify)
	}

	config = xpert.IBMWatsonMQTTConfig{ // nolint: staticcheck
		BaseMqttConfig: xpert.BaseMqttConfig{
			BrokerAddress: brokerAddr,
			ClientId:      clientID,
			Topic:         topic,
			AutoReconnect: autoReconnect,
			QoS:           byte(qos),
			Retain:        retain,
		},
		SecretPath:     secretPath,
		SkipCertVerify: skipCertVerify,
	}

	return config, persistOnError, nil
}
