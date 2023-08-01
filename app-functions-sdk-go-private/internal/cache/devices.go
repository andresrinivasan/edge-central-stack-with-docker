// Copyright (C) 2022 IOTech Ltd

package cache

import (
	"context"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/google/uuid"
)

var (
	initOnceForDeviceCache sync.Once
	dc                     *deviceCache
)

type DeviceCache interface {
	ForName(deviceName string) (dtos.Device, bool)
}

type deviceCache struct {
	deviceMap   map[string]dtos.Device
	absentDrMap map[string]bool
	appContext  interfaces.AppFunctionContext
	mutex       sync.RWMutex
}

func (d *deviceCache) ForName(deviceName string) (dtos.Device, bool) {
	device, ok, isIgnored := d.getCachedDevice(deviceName)

	// retrieve device through DeviceClient again
	if !ok && !isIgnored && d.appContext != nil {
		device, ok = d.getDevice(deviceName)
	}
	return device, ok
}

func (d *deviceCache) getCachedDevice(key string) (dtos.Device, bool, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, isIgnored := d.absentDrMap[key]
	device, ok := d.deviceMap[key]
	return device, ok, isIgnored
}

func (d *deviceCache) getDevice(deviceName string) (dtos.Device, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	device, err := d.appContext.XpertGetDevice(deviceName)
	ok := false
	if err != nil {
		// If app service receives EdgeX Events from outside, the corresponding deviceName
		// may not be found in Core Metadata. In this case, there is no need to query again in the future.
		// Otherwise the memory usage will climb up when such events arrive at high frequency.
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			d.absentDrMap[deviceName] = true
			d.appContext.LoggingClient().Warnf("the Device %s doesn't exist, "+
				"ignoring this query in the future", deviceName)
		} else {
			d.appContext.LoggingClient().Warnf("encountering error while retrieving Device. "+
				"Device Name: %s, Error: %v", deviceName, err)
		}
	} else {
		d.deviceMap[deviceName] = device
		ok = true
	}
	return device, ok
}

func Devices(appContext interfaces.AppFunctionContext) DeviceCache {
	if dc == nil {
		initOnceForDeviceCache.Do(func() {
			var deviceMap map[string]dtos.Device

			ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String()) //nolint: staticcheck
			res, err := appContext.DeviceClient().AllDevices(ctx, nil, 0, -1)
			if err != nil {
				appContext.LoggingClient().Warnf("Device cache initialization failed: %v", err)
			} else {
				deviceMap = make(map[string]dtos.Device, len(res.Devices))
				for _, d := range res.Devices {
					deviceMap[d.Name] = d
				}
			}
			dc = &deviceCache{deviceMap: deviceMap, absentDrMap: make(map[string]bool), appContext: appContext}
		})
	}
	return dc
}
