// Copyright (C) 2021-2022 IOTech Ltd

package xpert

import (
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"testing"

	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var existingTags = map[string]interface{}{
	"Tag1": "Value1",
	"Tag2": "Value2",
}

var tagsToAdd = map[string]interface{}{
	"latitude": map[string]interface{}{
		"degree": float64(25),
		"minute": float64(1),
		"second": float64(26.1528),
	},
	"longitude": map[string]interface{}{
		"degree": float64(121),
		"minute": float64(31),
		"second": float64(21.2478),
	},
}

var allTagsAdded = map[string]interface{}{
	"Tag1": "Value1",
	"Tag2": "Value2",
	"latitude": map[string]interface{}{
		"degree": float64(25),
		"minute": float64(1),
		"second": float64(26.1528),
	},
	"longitude": map[string]interface{}{
		"degree": float64(121),
		"minute": float64(31),
		"second": float64(21.2478),
	},
}

func TestXpertAddTagsFromDeviceResource(t *testing.T) {
	resourceName := "test"

	eventWithExistingTags := dtos.Event{
		Readings: []dtos.BaseReading{
			{ResourceName: resourceName},
		},
		Tags: existingTags,
	}
	eventWithoutTags := eventWithExistingTags
	eventWithoutTags.Tags = nil

	tests := []struct {
		Name          string
		FunctionInput interface{}
		Expected      map[string]interface{}
		ErrorExpected bool
		ErrorContains string
	}{
		{"No existing Event tags", eventWithoutTags, tagsToAdd, false, ""},
		{"Event has existing tags", eventWithExistingTags, allTagsAdded, false, ""},
		{"No tags added", eventWithExistingTags, eventWithExistingTags.Tags, false, ""},
		{"Error - No data", nil, nil, true, "no Event Received"},
		{"Error - Input not event", "Not an Event", nil, true, "not an Event"},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			var continuePipeline bool
			var result interface{}
			tags := NewTags()
			if testCase.FunctionInput != nil {
				continuePipeline, result = tags.AddTagsFromDeviceResource(ctx, testCase.FunctionInput)
			} else {
				continuePipeline, result = tags.AddTagsFromDeviceResource(ctx, nil)
			}

			if testCase.ErrorExpected {
				err := result.(error)
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ErrorContains)
				require.False(t, continuePipeline)
				return // Test completed
			}

			assert.True(t, continuePipeline)
			actual, ok := result.(dtos.Event)
			require.True(t, ok, "Result not an Event")
			assert.Equal(t, testCase.Expected, actual.Tags)
		})
	}
}

func TestXpertAddTagsFromDevice(t *testing.T) {
	deviceName := "test-device"

	mockClient := clientMocks.DeviceClient{}
	mockClient.On("AllDevices", mock.Anything, mock.Anything, 0, -1).Return(responses.MultiDevicesResponse{
		BaseWithTotalCountResponse: common.BaseWithTotalCountResponse{},
		Devices: []dtos.Device{{

			Name: deviceName,
			Tags: tagsToAdd,
		}},
	}, nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return &mockClient
		},
	})

	eventWithExistingTags := dtos.Event{
		DeviceName: deviceName,
		Tags:       existingTags,
	}
	eventWithoutTags := eventWithExistingTags
	eventWithoutTags.Tags = nil

	tests := []struct {
		Name          string
		FunctionInput interface{}
		Expected      map[string]interface{}
		ErrorExpected bool
		ErrorContains string
	}{
		{"No existing Event tags", eventWithoutTags, tagsToAdd, false, ""},
		{"Event has existing tags", eventWithExistingTags, allTagsAdded, false, ""},
		{"No tags added", eventWithExistingTags, existingTags, false, ""},
		{"Error - No data", nil, nil, true, "no Event received"},
		{"Error - Input not event", "Not an Event", nil, true, "not an Event"},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			var continuePipeline bool
			var result interface{}
			tags := NewTags()
			if testCase.FunctionInput != nil {
				continuePipeline, result = tags.AddTagsFromDevice(ctx, testCase.FunctionInput)
			} else {
				continuePipeline, result = tags.AddTagsFromDevice(ctx, nil)
			}

			if testCase.ErrorExpected {
				err := result.(error)
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ErrorContains)
				require.False(t, continuePipeline)
				return // Test completed
			}

			assert.True(t, continuePipeline)
			actual, ok := result.(dtos.Event)
			require.True(t, ok, "Result not an Event")
			assert.Equal(t, testCase.Expected, actual.Tags)
		})
	}
}
