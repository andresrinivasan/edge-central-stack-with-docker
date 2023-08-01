// Copyright (C) 2021-2023 IOTech Ltd

package xpert

import (
	"errors"
	"fmt"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	gometrics "github.com/rcrowley/go-metrics"
)

type MqttClientFactory interface {
	Create(ctx interfaces.AppFunctionContext) (MQTT.Client, error)
}

// BaseMqttConfig represents the common configuration for a MQTT client
type BaseMqttConfig struct {
	// BrokerAddress should be set to the complete broker address i.e. mqtts://mosquitto:8883/mybroker
	BrokerAddress string
	// ClientId to connect with the broker with.
	ClientId string
	// AutoReconnect indicated whether or not to retry connection if disconnected
	AutoReconnect bool
	// Topic that you wish to publish to
	Topic string
	// QoS for MQTT Connection
	QoS byte
	// Retain setting for MQTT Connection
	Retain bool
}

// BaseMqttSender is the base MQTT client that could be embedded into MqttSenders of other Cloud vendors
// BaseMqttSender implements common codes to deal with Mqtt connection and publish
type BaseMqttSender struct {
	config               BaseMqttConfig
	client               MQTT.Client
	persistOnError       bool
	secretsLastRetrieved time.Time
	mqttClientFactory    MqttClientFactory
	usingSharedClient    bool
	mqttSizeMetrics      gometrics.Histogram
}

func (sender *BaseMqttSender) initializeMQTTClient(ctx interfaces.AppFunctionContext) error {
	ctx.SharedMQTTClientMutex().Lock()
	defer ctx.SharedMQTTClientMutex().Unlock()

	if sender.usingSharedClient {
		sender.client, sender.secretsLastRetrieved = ctx.SharedMQTTClient().Get()
	}

	// If the conditions changed while waiting for the lock, i.e. other thread completed the initialization,
	// then skip doing anything
	if sender.client != nil && !sender.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		return nil
	}

	client, err := sender.mqttClientFactory.Create(ctx)
	if err != nil {
		return err
	}

	sender.client = client
	sender.secretsLastRetrieved = time.Now()

	if sender.usingSharedClient {
		ctx.SharedMQTTClient().Set(sender.client, sender.secretsLastRetrieved)
	}

	return nil
}

func (sender *BaseMqttSender) connect(ctx interfaces.AppFunctionContext) error {
	// note that each individual message will be processed in a separate goroutine, so under the circumstance where
	// target mqtt broker is unreachable (e.g. incorrect hostname) and messages arrives in high frequency(e.g. 5 messages
	// per 200ms), there will be lots of goroutines attempting to make connection to MQTT broker.  To avoid abnormal
	// memory soaring caused by too many goroutines queuing to acquire ctx.SharedMQTTClientMutex().Lock(), we use
	// ctx.MqttConnectionWaitingCounter() here to ensure only certain number of goroutines can be allowed to acquire
	// ctx.SharedMQTTClientMutex().Lock()
	err := ctx.MqttConnectionWaitingCounter().Add(1)
	if err != nil {
		return err
	}
	ctx.SharedMQTTClientMutex().Lock()
	defer func() {
		ctx.MqttConnectionWaitingCounter().Subtract(1)
		ctx.SharedMQTTClientMutex().Unlock()
	}()

	// check if mqtt client is initialized or not
	if sender.client == nil {
		return fmt.Errorf("mqtt client is not initialized")
	}
	// If other thread made the connection while this one was waiting for the lock
	// then skip trying to connect
	if sender.client.IsConnected() {
		return nil
	}

	token := sender.client.Connect()
	if !token.WaitTimeout(WaitDuration) {
		return fmt.Errorf("connection timeout while connecting to MQTT Broker: %s", sender.config.BrokerAddress)
	}
	if token.Error() != nil {
		return fmt.Errorf("could not connect to MQTT Broker: %s. Error: %v", sender.config.BrokerAddress, token.Error())
	}
	ctx.LoggingClient().Infof("Successfully connected to MQTT Broker: %s", sender.config.BrokerAddress)
	return nil
}

func (sender *BaseMqttSender) publish(ctx interfaces.AppFunctionContext, exportData []byte) error {
	if sender.client == nil {
		return fmt.Errorf("mqtt client is not initialized")
	}
	if !sender.client.IsConnected() {
		err := sender.connect(ctx)
		if err != nil {
			sender.setRetryData(ctx, exportData)
			subMessage := "dropping event"
			if sender.persistOnError {
				subMessage = "persisting Event for later retry"
			}
			ctx.LoggingClient().Errorf("failed to connect to MQTT broker, %s. Error: ", subMessage, err)
			return err
		}
	}
	// replace the placeholders of the form '{any-value-key}' in the topic with the value stored in AppFunctionContext
	topic, err := ctx.ApplyValues(sender.config.Topic)
	if err != nil {
		return err
	}
	token := sender.client.Publish(topic, sender.config.QoS, sender.config.Retain, exportData)
	if !token.WaitTimeout(WaitDuration) {
		sender.setRetryData(ctx, exportData)
		return fmt.Errorf("connection timeout while publishing to MQTT Broker: %v", sender.config.BrokerAddress)
	}
	if token.Error() != nil {
		sender.setRetryData(ctx, exportData)
		return fmt.Errorf("failed to publish data to MQTT Broker: %s.  Error:%v", sender.config.BrokerAddress, token.Error())
	}

	// capture the size for metrics
	exportDataBytes := len(exportData)
	if sender.mqttSizeMetrics == nil {
		var err error
		tag := fmt.Sprintf("%s/%s", sender.config.BrokerAddress, topic)
		metricName := fmt.Sprintf("%s-%s", internal.MqttExportSizeName, tag)
		ctx.LoggingClient().Debugf("Initializing metric %s.", metricName)
		sender.mqttSizeMetrics = gometrics.NewHistogram(gometrics.NewUniformSample(internal.MetricsReservoirSize))
		metricsManger := ctx.MetricsManager()
		if metricsManger != nil {
			err = metricsManger.Register(metricName, sender.mqttSizeMetrics, map[string]string{"address/topic": tag})
		} else {
			err = errors.New("metrics manager not available")
		}
		if err != nil {
			ctx.LoggingClient().Errorf("Unable to register metric %s. Collection will continue, but metric will not be reported: %s", internal.MqttExportSizeName, err.Error())
		}
	}
	sender.mqttSizeMetrics.Update(int64(exportDataBytes))
	ctx.LoggingClient().Debugf("Sent %d bytes of data to MQTT Broker in pipeline '%s'", exportDataBytes, ctx.PipelineId())
	ctx.LoggingClient().Tracef("Data exported", "Transport", "MQTT", "pipeline", ctx.PipelineId(), common.CorrelationHeader, ctx.CorrelationID())

	return nil
}

func (sender *BaseMqttSender) setRetryData(ctx interfaces.AppFunctionContext, exportData []byte) {
	if sender.persistOnError {
		ctx.SetRetryData(exportData)
	}
}
