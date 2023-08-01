// Copyright (C) 2022 IOTech Ltd

package cache

import (
	"context"
	"crypto/md5" // #nosec
	"encoding/json"
	"fmt"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	valueTypeIdBool = iota + 1
	valueTypeIdString
	valueTypeIdUint8
	valueTypeIdUint16
	valueTypeIdUint32
	valueTypeIdUint64
	valueTypeIdInt8
	valueTypeIdInt16
	valueTypeIdInt32
	valueTypeIdInt64
	valueTypeIdFloat32
	valueTypeIdFloat64
	valueTypeIdBinary
	valueTypeIdObject
	valueTypeIdBoolArray
	valueTypeIdStringArray
	valueTypeIdUint8Array
	valueTypeIdUint16Array
	valueTypeIdUint32Array
	valueTypeIdUint64Array
	valueTypeIdInt8Array
	valueTypeIdInt16Array
	valueTypeIdInt32Array
	valueTypeIdInt64Array
	valueTypeIdFloat32Array
	valueTypeIdFloat64Array
)

var (
	pdrc                *postgresDeviceResourceCache
	pvtc                *postgresValueTypeCache
	pdrcInitOnce        sync.Once
	pvtcInitOnce        sync.Once
	HardcodedValueTypes = map[int16]string{
		valueTypeIdBool:         common.ValueTypeBool,
		valueTypeIdString:       common.ValueTypeString,
		valueTypeIdUint8:        common.ValueTypeUint8,
		valueTypeIdUint16:       common.ValueTypeUint16,
		valueTypeIdUint32:       common.ValueTypeUint32,
		valueTypeIdUint64:       common.ValueTypeUint64,
		valueTypeIdInt8:         common.ValueTypeInt8,
		valueTypeIdInt16:        common.ValueTypeInt16,
		valueTypeIdInt32:        common.ValueTypeInt32,
		valueTypeIdInt64:        common.ValueTypeInt64,
		valueTypeIdFloat32:      common.ValueTypeFloat32,
		valueTypeIdFloat64:      common.ValueTypeFloat64,
		valueTypeIdBinary:       common.ValueTypeBinary,
		valueTypeIdObject:       common.ValueTypeObject,
		valueTypeIdBoolArray:    common.ValueTypeBoolArray,
		valueTypeIdStringArray:  common.ValueTypeStringArray,
		valueTypeIdUint8Array:   common.ValueTypeUint8Array,
		valueTypeIdUint16Array:  common.ValueTypeUint16Array,
		valueTypeIdUint32Array:  common.ValueTypeUint32Array,
		valueTypeIdUint64Array:  common.ValueTypeUint64Array,
		valueTypeIdInt8Array:    common.ValueTypeInt8Array,
		valueTypeIdInt16Array:   common.ValueTypeInt16Array,
		valueTypeIdInt32Array:   common.ValueTypeInt32Array,
		valueTypeIdInt64Array:   common.ValueTypeInt64Array,
		valueTypeIdFloat32Array: common.ValueTypeFloat32Array,
		valueTypeIdFloat64Array: common.ValueTypeFloat64Array,
	}
)

type PostgresDeviceResourceCache interface {
	Get(deviceName, resourceName string, tags map[string]interface{}) (PostgresDeviceResource, bool)
	Add(id int32, reading dtos.BaseReading, tags map[string]interface{}) bool
}

type PostgresValueTypeCache interface {
	Get(valueTypeName string) (PostgresValueType, bool)
}

type PostgresDeviceResource struct {
	Id           int32
	DeviceName   string
	ResourceName string
	Tags         map[string]interface{}
}

type PostgresValueType struct {
	Id            int16
	ValueTypeName string
}

type postgresDeviceResourceCache struct {
	drMap      map[string]PostgresDeviceResource // key is device name to concatenate resource name and tags
	appContext interfaces.AppFunctionContext
	mutex      sync.RWMutex
}

type postgresValueTypeCache struct {
	vtMap      map[string]PostgresValueType // key is value type name
	appContext interfaces.AppFunctionContext
	mutex      sync.RWMutex
}

func deviceResourceCacheKey(deviceName, resourceName string, tags map[string]interface{}) (string, error) {
	jsonEncodedTags, err := json.Marshal(tags)
	if err != nil {
		return "", fmt.Errorf("failed to encode tags to JSON, err: %v", err)
	}
	byteData := append([]byte(deviceName), []byte(resourceName)...)
	byteData = append(byteData, jsonEncodedTags...)
	return fmt.Sprintf("%x", md5.Sum(byteData)), nil // #nosec
}

func (d *postgresValueTypeCache) Get(valueTypeName string) (PostgresValueType, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	dr, ok := d.vtMap[valueTypeName]
	return dr, ok
}

func (d *postgresDeviceResourceCache) Get(deviceName, resourceName string, tags map[string]interface{}) (PostgresDeviceResource, bool) {
	key, err := deviceResourceCacheKey(deviceName, resourceName, tags)
	if err != nil {
		d.appContext.LoggingClient().Error(err.Error())
		return PostgresDeviceResource{}, false
	}
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	dr, ok := d.drMap[key]
	return dr, ok
}

func (d *postgresDeviceResourceCache) Add(id int32, reading dtos.BaseReading, tags map[string]interface{}) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	key, err := deviceResourceCacheKey(reading.DeviceName, reading.ResourceName, tags)
	if err != nil {
		d.appContext.LoggingClient().Error(err.Error())
		return false
	}
	d.drMap[key] = PostgresDeviceResource{
		Id:           id,
		DeviceName:   reading.DeviceName,
		ResourceName: reading.ResourceName,
		Tags:         tags,
	}
	return true
}

func PostgresValueTypes(appContext interfaces.AppFunctionContext, pgxConn *pgxpool.Conn, sqlQuery string) PostgresValueTypeCache {
	if pvtc != nil {
		return pvtc
	}
	pvtcInitOnce.Do(func() {
		vtMap := make(map[string]PostgresValueType)
		rows, err := pgxConn.Query(context.Background(), sqlQuery)
		if err != nil {
			appContext.LoggingClient().Errorf("Postgres ValueType cache initialization failed: %v", err)
		} else {
			for rows.Next() {
				valueType := PostgresValueType{}
				if err := rows.Scan(&valueType.Id, &valueType.ValueTypeName); err != nil {
					appContext.LoggingClient().Errorf("failed to read value from the query result, err: %v", err)
				}
				vtMap[valueType.ValueTypeName] = valueType
			}
		}
		pvtc = &postgresValueTypeCache{vtMap: vtMap, appContext: appContext}
	})
	return pvtc
}

func PostgresDeviceResources(appContext interfaces.AppFunctionContext, pgxConn *pgxpool.Conn, sqlQuery string) PostgresDeviceResourceCache {
	if pdrc != nil {
		return pdrc
	}
	pdrcInitOnce.Do(func() {
		drMap := make(map[string]PostgresDeviceResource)
		rows, err := pgxConn.Query(context.Background(), sqlQuery)
		if err != nil {
			appContext.LoggingClient().Errorf("Postgres Device Resource cache initialization failed: %v", err)
		} else {
			for rows.Next() {
				resource := PostgresDeviceResource{}
				if err := rows.Scan(&resource.Id, &resource.DeviceName, &resource.ResourceName, &resource.Tags); err != nil {
					appContext.LoggingClient().Errorf("failed to read value from the query result, err: %v", err)
				}
				key, err := deviceResourceCacheKey(resource.DeviceName, resource.ResourceName, resource.Tags)
				if err != nil {
					appContext.LoggingClient().Error(err.Error())
				}
				drMap[key] = resource
			}
		}
		pdrc = &postgresDeviceResourceCache{drMap: drMap, appContext: appContext}
	})
	return pdrc
}
