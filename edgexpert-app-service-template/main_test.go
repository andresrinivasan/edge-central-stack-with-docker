// TODO: Change Copyright to your company if open sourcing or remove header
//
// Copyright (C) 2021-2022 IOTech Ltd

package main

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// This is an example of how to test the code that would typically be in the main() function use mocks
// Not to helpful for a simple main() , but can be if the main() has more complexity that should be unit tested
// TODO: add/update tests for your customized CreateAndRunAppService or remove if your main code doesn't require unit testing.

func TestCreateAndRunService_Success(t *testing.T) {
	app := myApp{}

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return([]string{"Random-Boolean-Device, Random-Integer-Device"}, nil)
		mockAppService.On("SetDefaultFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("AddFunctionsPipelineForTopics", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("LoadCustomConfig", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
			// set the required configuration so validation passes
			app.serviceConfig.AppCustom.SomeValue = 987
			app.serviceConfig.AppCustom.SomeService.Host = "SomeHost"
		})
		mockAppService.On("ListenForCustomConfigChanges", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("MakeItRun").Return(nil)

		return mockAppService, true
	}

	expected := 0
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_NewService_Failed(t *testing.T) {
	app := myApp{}

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		return nil, false
	}
	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_GetAppSettingStrings_Failed(t *testing.T) {
	app := myApp{}

	getAppSettingStringsCalled := false
	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return(nil, fmt.Errorf("Failed")).Run(func(args mock.Arguments) {
			getAppSettingStringsCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	require.True(t, getAppSettingStringsCalled, "GetAppSettingStrings never called")
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_SetFunctionsPipeline_Failed(t *testing.T) {
	app := myApp{}

	// ensure failure is from SetFunctionsPipeline
	setFunctionsPipelineCalled := false

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return([]string{"Random-Boolean-Device, Random-Integer-Device"}, nil)
		mockAppService.On("LoadCustomConfig", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
			// set the required configuration so validation passes
			app.serviceConfig.AppCustom.SomeValue = 987
			app.serviceConfig.AppCustom.SomeService.Host = "SomeHost"
		})
		mockAppService.On("ListenForCustomConfigChanges", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("SetDefaultFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(fmt.Errorf("Failed")).Run(func(args mock.Arguments) {
			setFunctionsPipelineCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	require.True(t, setFunctionsPipelineCalled, "SetFunctionsPipeline never called")
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_MakeItRun_Failed(t *testing.T) {
	app := myApp{}

	// ensure failure is from MakeItRun
	makeItRunCalled := false

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return([]string{"Random-Boolean-Device, Random-Integer-Device"}, nil)
		mockAppService.On("LoadCustomConfig", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
			// set the required configuration so validation passes
			app.serviceConfig.AppCustom.SomeValue = 987
			app.serviceConfig.AppCustom.SomeService.Host = "SomeHost"
		})
		mockAppService.On("ListenForCustomConfigChanges", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("SetDefaultFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("AddFunctionsPipelineForTopics", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("MakeItRun").Return(fmt.Errorf("Failed")).Run(func(args mock.Arguments) {
			makeItRunCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	require.True(t, makeItRunCalled, "MakeItRun never called")
	assert.Equal(t, expected, actual)
}
