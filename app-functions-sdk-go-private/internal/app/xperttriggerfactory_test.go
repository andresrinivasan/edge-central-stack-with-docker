// Copyright (C) 2021 IOTech Ltd

package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mqtt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTriggers(t *testing.T) {
	customType := "Telegraph"
	triggerTypes := []string{
		internal.TriggerTypeHTTP,
		internal.TriggerTypeMessageBus,
		internal.TriggerTypeMQTT,
		customType,
	}
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: strings.Join(triggerTypes, ","),
			},
		},
		runtime: runtime.NewGolangRuntime("", nil, dic),
		dic:     dic,
		lc:      logger.MockLogger{},
	}

	err := sdk.RegisterCustomTriggerFactory(customType, func(c interfaces.TriggerConfig) (interfaces.Trigger, error) {
		return &mockCustomTrigger{}, nil
	})
	require.NoError(t, err)

	triggers, err := sdk.setupTriggers()
	require.NoError(t, err)
	require.NotNil(t, triggers, "triggers should not be nil")

	for _, triggerType := range triggerTypes {
		switch triggerType {
		case internal.TriggerTypeHTTP:
			assert.IsType(t, &http.Trigger{}, triggers[triggerType], "should be an http trigger")
		case internal.TriggerTypeMessageBus:
			assert.IsType(t, &messagebus.Trigger{}, triggers[triggerType], "should be an edgex-messagebus trigger")
		case internal.TriggerTypeMQTT:
			assert.IsType(t, &mqtt.Trigger{}, triggers[triggerType], "should be an external-MQTT trigger")
		case customType:
			assert.IsType(t, &mockCustomTrigger{}, triggers[strings.ToUpper(customType)], "should be a custom trigger")
		}
	}
}

func TestSetupTriggersNoTypesDefined(t *testing.T) {
	sdk := Service{
		config:  &common.ConfigurationStruct{},
		runtime: runtime.NewGolangRuntime("", nil, dic),
		lc:      logger.MockLogger{},
		dic:     dic,
	}

	triggers, err := sdk.setupTriggers()
	require.Error(t, err)
	assert.Equal(t, err.Error(), "no Trigger Type defined")
	require.Nil(t, triggers, "triggers should be nil")
}

func TestSetupTriggersUnregisteredCustomTrigger(t *testing.T) {
	customType := "Telegraph"
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: customType,
			},
		},
		runtime: runtime.NewGolangRuntime("", nil, dic),
		lc:      logger.MockLogger{},
		dic:     dic,
	}

	triggers, err := sdk.setupTriggers()
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("[%s] is not registered", strings.ToUpper(customType)))
	require.Nil(t, triggers, "triggers should be nil")
}

func TestSetupTriggersCustomTriggerBuildError(t *testing.T) {
	customType := "Telegraph"
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: customType,
			},
		},
		runtime: runtime.NewGolangRuntime("", nil, dic),
		lc:      logger.MockLogger{},
		dic:     dic,
	}

	expectedErrMsg := "failed to build custom trigger"
	err := sdk.RegisterCustomTriggerFactory(customType, func(c interfaces.TriggerConfig) (interfaces.Trigger, error) {
		return &mockCustomTrigger{}, errors.New(expectedErrMsg)
	})
	require.NoError(t, err)

	triggers, err := sdk.setupTriggers()
	require.Error(t, err)
	assert.Equal(t, err.Error(), expectedErrMsg)
	require.Nil(t, triggers, "triggers should be nil")
}
