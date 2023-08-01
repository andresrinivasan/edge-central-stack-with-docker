// Copyright (C) 2021-2023 IOTech Ltd

package interfaces

import (
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common/xpert"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type XpertAppFunctionContext interface {
	// SharedMQTTClientMutex returns the reference to SDK runtime's sharedMQTTClientMutex, which is used to prevent
	// concurrent processes from entering the section of code used to initialize the external MQTT client and create connection.
	SharedMQTTClientMutex() *sync.Mutex
	// SetSharedMQTTClientMutex sets the reference to the sharedMQTTClientMutex of SDK runtime.
	SetSharedMQTTClientMutex(mutex *sync.Mutex)
	// SharedMQTTClient returns the shared MQTT client for subscribing/publishing to an external MQTT broker.
	SharedMQTTClient() *xpert.SharedMQTTClient
	// SetSharedMQTTClient sets the reference to the SharedMQTTClient of SDK runtime.
	SetSharedMQTTClient(client *xpert.SharedMQTTClient)
	// XpertGetDevice returns the Device by the given device name
	XpertGetDevice(deviceName string) (dtos.Device, errors.EdgeX)
	// XpertGetDeviceResource retrieves the DeviceResource for given profileName and resourceName.
	// Resources retrieved are cached so multiple calls for same profileName and resourceName don't result in multiple
	// unneeded HTTP calls to Core Metadata
	// The implementation of XpertGetDeviceResource is almost identical to GetDeviceResource except for the error type.
	// XpertGetDeviceResource returns an EdgeX Error to the calling function for better error handling.
	XpertGetDeviceResource(profileName string, resourceName string) (dtos.DeviceResource, errors.EdgeX)
	// MqttConnectionWaitingCounter returns the single AtomicCounter of SDK runtime for concurrent BaseMqttSender.connect
	// to share number of goroutines attempting to make first connection.
	MqttConnectionWaitingCounter() *xpert.AtomicCounter
	// SetMqttConnectionWaitingCounter passes the the single AtomicCounter of SDK runtime into AppFunctionContext.
	SetMqttConnectionWaitingCounter(*xpert.AtomicCounter)
}
