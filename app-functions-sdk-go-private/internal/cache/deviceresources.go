// Copyright (C) 2020-2022 IOTech Ltd

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
	initOnce sync.Once
	drc      *deviceResourceCache
)

type DeviceResourceCache interface {
	ForName(profileName, resourceName string) (dtos.DeviceResource, bool)
}

type deviceResourceCache struct {
	drMap       map[string]dtos.DeviceResource // key is DeviceProfile name to concatenate DeviceResource name
	absentDrMap map[string]bool                // key is DeviceProfile name to concatenate DeviceResource name
	appContext  interfaces.AppFunctionContext
	mutex       sync.RWMutex
}

func (d *deviceResourceCache) ForName(profileName, resourceName string) (dtos.DeviceResource, bool) {
	dr, ok, isIgnored := d.getCachedDeviceResource(profileName + resourceName)

	// retrieve device resource through DeviceProfileClient again
	if !ok && !isIgnored && d.appContext != nil {
		dr, ok = d.getDeviceResource(profileName, resourceName)
	}
	return dr, ok
}

func (d *deviceResourceCache) getCachedDeviceResource(key string) (dtos.DeviceResource, bool, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, isIgnored := d.absentDrMap[key]
	dr, ok := d.drMap[key]
	return dr, ok, isIgnored
}

func (d *deviceResourceCache) getDeviceResource(profileName, resourceName string) (dtos.DeviceResource, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	deviceResource, err := d.appContext.XpertGetDeviceResource(profileName, resourceName)
	ok := false
	if err != nil {
		// If app service receives EdgeX Events from outside, the corresponding profileName and resourceName
		// may not be found in Core Metadata. In this case, there is no need to query again in the future.
		// Otherwise the memory usage will climb up when such events arrive at high frequency.
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			d.absentDrMap[profileName+resourceName] = true
			d.appContext.LoggingClient().Warnf("the DeviceResource %s doesn't exist in Device Profile %s, "+
				"ignoring this query in the future", resourceName, profileName)
		} else {
			d.appContext.LoggingClient().Warnf("encountering error while retrieving DeviceResource. "+
				"Profile Name: %s, Resource Name: %s. Error: %v", profileName, resourceName, err)
		}
	} else {
		ok = true
		d.drMap[profileName+resourceName] = deviceResource
	}
	return deviceResource, ok
}

func DeviceResources(appContext interfaces.AppFunctionContext) DeviceResourceCache {
	if drc == nil {
		initOnce.Do(func() {
			var drMap map[string]dtos.DeviceResource
			ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String()) //nolint: staticcheck
			res, err := appContext.DeviceProfileClient().AllDeviceProfiles(ctx, nil, 0, -1)
			if err != nil {
				appContext.LoggingClient().Warnf("Device Resource cache initialization failed: %v", err)
				drMap = make(map[string]dtos.DeviceResource)
			} else {
				defaultSize := 0
				for _, p := range res.Profiles {
					defaultSize += len(p.DeviceResources)
				}
				drMap = make(map[string]dtos.DeviceResource, defaultSize*2)
				for _, p := range res.Profiles {
					for _, r := range p.DeviceResources {
						drMap[p.Name+r.Name] = r
					}
				}
			}
			drc = &deviceResourceCache{drMap: drMap, absentDrMap: make(map[string]bool), appContext: appContext}
		})
	}
	return drc
}
