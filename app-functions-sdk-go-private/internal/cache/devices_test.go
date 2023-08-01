// Copyright (C) 2022 IOTech Ltd

package cache

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/stretchr/testify/assert"
)

var (
	TestDevice = "testDevice"
	tagsToAdd  = map[string]interface{}{
		"tag1": "value1",
		"tag2": "value2",
	}
)

func Test_deviceCache_ForName(t *testing.T) {
	dc = &deviceCache{
		deviceMap: map[string]dtos.Device{TestDevice: {Name: TestDevice, Tags: tagsToAdd}},
	}

	testDevice := dtos.Device{
		Name: TestDevice,
		Tags: tagsToAdd,
	}
	tests := []struct {
		name       string
		deviceName string
		device     dtos.Device
		expected   bool
	}{
		{"Invalid - empty name", "", dtos.Device{}, false},
		{"Invalid - nonexistent Device name", "nil", dtos.Device{}, false},
		{"Valid", TestDevice, testDevice, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := dc.ForName(tt.deviceName)
			assert.Equal(t, res, tt.device, "ForName returns wrong Device")
			assert.Equal(t, ok, tt.expected, "ForName returns opposite result")
		})
	}
}
