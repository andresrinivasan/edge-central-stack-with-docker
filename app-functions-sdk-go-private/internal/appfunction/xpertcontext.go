// Copyright (C) 2021-2023 IOTech Ltd

package appfunction

import (
	"context"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common/xpert"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type XpertContext struct {
	// sharedMQTTClient is a pointer to the sharedMQTTClient held by GolangRuntime.
	// sharedMQTTclient will be initialized just once and can be reused by pipeline triggers and pipeline functions
	// to connect to an external MQTT broker, its connection information is specified in the [Trigger.ExternalMqtt] section of service configuration.
	sharedMQTTClient *xpert.SharedMQTTClient //nolint: structcheck
	// sharedMQTTClientMutex is a pointer to the sharedMQTTClientMutex held by GolangRuntime, which is used to make sure
	// only one goroutine can access sharedMQTTClient.
	sharedMQTTClientMutex *sync.Mutex //nolint: structcheck
	// mqttConnectionWaitingCounter is a AtomicCounter for concurrent BaseMqttSender.connect to share number of
	// goroutines attempting to make first connection.
	mqttConnectionWaitingCounter *xpert.AtomicCounter
}

func (appContext *Context) SharedMQTTClientMutex() *sync.Mutex {
	return appContext.sharedMQTTClientMutex
}

func (appContext *Context) SetSharedMQTTClientMutex(mutex *sync.Mutex) {
	appContext.sharedMQTTClientMutex = mutex
}

func (appContext *Context) SharedMQTTClient() *xpert.SharedMQTTClient {
	return appContext.sharedMQTTClient
}

func (appContext *Context) SetSharedMQTTClient(client *xpert.SharedMQTTClient) {
	appContext.sharedMQTTClient = client
}

func (appContext *Context) MqttConnectionWaitingCounter() *xpert.AtomicCounter {
	return appContext.mqttConnectionWaitingCounter
}

func (appContext *Context) SetMqttConnectionWaitingCounter(counter *xpert.AtomicCounter) {
	appContext.mqttConnectionWaitingCounter = counter
}

// XpertGetDevice returns the Device by the given device name
func (appContext *Context) XpertGetDevice(deviceName string) (dtos.Device, edgexErr.EdgeX) {
	client := appContext.DeviceClient()
	if client == nil {
		return dtos.Device{}, edgexErr.NewCommonEdgeX(edgexErr.KindContractInvalid,
			"DeviceClient not initialized. Core Metadata is missing from clients configuration", nil)
	}

	response, err := client.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return dtos.Device{}, err
	}

	return response.Device, nil
}

// XpertGetDeviceResource retrieves the DeviceResource for given profileName and resourceName, and the difference
// between the counterpart in contex.go is the error type. We need errors.EdgeX to identify the error kind.
func (appContext *Context) XpertGetDeviceResource(profileName string, resourceName string) (dtos.DeviceResource, edgexErr.EdgeX) {
	client := appContext.DeviceProfileClient()
	if client == nil {
		return dtos.DeviceResource{}, edgexErr.NewCommonEdgeX(edgexErr.KindContractInvalid,
			"DeviceProfileClient not initialized. Core Metadata is missing from clients configuration", nil)
	}

	response, err := client.DeviceResourceByProfileNameAndResourceName(context.Background(), profileName, resourceName)
	if err != nil {
		return dtos.DeviceResource{}, err
	}

	return response.Resource, nil
}
