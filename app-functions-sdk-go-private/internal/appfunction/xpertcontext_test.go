// Copyright (C) 2021-2022 IOTech Ltd

package appfunction

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common/xpert"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestContext_SetExternalMQTTClientMutex(t *testing.T) {
	expected := &sync.Mutex{}

	target.SetSharedMQTTClientMutex(expected)
	actual := target.sharedMQTTClientMutex

	assert.Equal(t, expected, actual)
}

func TestContext_ExternalMQTTClientMutex(t *testing.T) {
	expected := &sync.Mutex{}

	target.sharedMQTTClientMutex = expected
	actual := target.SharedMQTTClientMutex()

	assert.Equal(t, expected, actual)
}

func TestContext_SetExternalMQTTClient(t *testing.T) {
	expected := &xpert.SharedMQTTClient{}

	target.SetSharedMQTTClient(expected)
	actual := target.sharedMQTTClient

	assert.Equal(t, expected, actual)
}

func TestContext_ExternalMQTTClient(t *testing.T) {
	expected := &xpert.SharedMQTTClient{}

	target.sharedMQTTClient = expected
	actual := target.SharedMQTTClient()

	assert.Equal(t, expected, actual)
}

func TestContext_GetDevice(t *testing.T) {
	deviceName := "test-device"
	device := dtos.Device{

		Name: deviceName,
	}
	mockClient := clientMocks.DeviceClient{}
	mockClient.On("DeviceByName", mock.Anything, deviceName).Return(responses.DeviceResponse{
		Device: device,
	}, nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return &mockClient
		},
	})

	actual, err := target.XpertGetDevice(deviceName)
	assert.Equal(t, device, actual)
	assert.NoError(t, err)
}
