// Copyright (C) 2021-2022 IOTech Ltd

package xpert

import (
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

type Tags struct{}

// NewTags creates, initializes and returns a new instance of Tags
func NewTags() Tags {
	return Tags{}
}

// AddTagsFromDeviceResource adds the pre-configured tags of Device Resource to the Event's tags collection.
func (t *Tags) AddTagsFromDeviceResource(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	ctx.LoggingClient().Debug("Adding tags from device resource to Event")

	if data == nil {
		return false, errors.New("no Event Received")
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	if event.Tags == nil {
		event.Tags = make(map[string]interface{})
	}

	for _, reading := range event.Readings {
		// get the DeviceResource from cache
		drc, ok := cache.DeviceResources(ctx).ForName(reading.ProfileName, reading.ResourceName)
		if !ok {
			ctx.LoggingClient().Warnf("No corresponding DeviceResource for reading %s.  Ignore by AddTagsFromDeviceResource", reading.Id)
		} else {
			for k, v := range drc.Tags {
				event.Tags[k] = v
			}
		}
	}

	if len(event.Tags) > 0 {
		ctx.LoggingClient().Debugf("Tags added to Event. Event tags=%v", event.Tags)
	} else {
		ctx.LoggingClient().Debug("No tags can be found in Device Resource.")
	}

	return true, event
}

// AddTagsFromDevice adds the pre-configured tags of Device to the Event's tags collection.
func (t *Tags) AddTagsFromDevice(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	ctx.LoggingClient().Debug("Adding tags from Device to Event")

	if data == nil {
		return false, errors.New("no Event received")
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("data type received is not an Event")
	}

	if event.Tags == nil {
		event.Tags = make(map[string]interface{})
	}

	// get the Device from cache
	dc, ok := cache.Devices(ctx).ForName(event.DeviceName)
	if !ok {
		ctx.LoggingClient().Warnf("No corresponding Device for Event %s.  Ignore by AddTagsFromDevice", event.Id)
	} else {
		if dc.Tags == nil {
			ctx.LoggingClient().Debug("No tags can be found in Device instance name: %s.", dc.Name)
		} else {
			for k, v := range dc.Tags {
				event.Tags[k] = v
				ctx.LoggingClient().Debugf("Event tag key=%v, value=%v added to Event from Device instance.", k, v)
			}
		}
	}

	return true, event
}
