// Copyright (C) 2021-2023 IOTech Ltd

package app

import (
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"

	"github.com/stretchr/testify/assert"
)

func TestConfigurableInfluxDBSyncWrite(t *testing.T) {
	configurable := Configurable{lc: lc}

	params := make(map[string]string)

	// no mandatory configuration specified in params
	trx := configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// influxdb server address provided
	params[InfluxDBServerURL] = "http://localhost:8086"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// unsupported authmode
	params[AuthMode] = "unsupported"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// supported authmode none
	params[AuthMode] = messaging.AuthModeNone
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// supported authmode token but no secretpath
	params[AuthMode] = xpert.InfluxDBAuthModeToken
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// supported authmode token with secretpath
	params[SecretPath] = "influxdb"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// unsupported valueType
	params[InfluxDBValueType] = "test"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// supported valueType float
	params[InfluxDBValueType] = xpert.InfluxDBValueTypeFloat
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// supported valueType integer
	params[InfluxDBValueType] = xpert.InfluxDBValueTypeInteger
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// supported valueType uinteger
	params[InfluxDBValueType] = xpert.InfluxDBValueTypeUInteger
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// supported valueType string
	params[InfluxDBValueType] = xpert.InfluxDBValueTypeString
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// supported valueType boolean
	params[InfluxDBValueType] = xpert.InfluxDBValueTypeBoolean
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// unparsible precision
	params[InfluxDBPrecision] = "test"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// parsible precision
	params[InfluxDBPrecision] = xpert.InfluxDBPrecisionNanoSeconds
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// parsible precision
	params[InfluxDBPrecision] = xpert.InfluxDBPrecisionMicroSeconds
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// parsible precision
	params[InfluxDBPrecision] = xpert.InfluxDBPrecisionMillieSeconds
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// parsible precision
	params[InfluxDBPrecision] = xpert.InfluxDBPrecisionSeconds
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite should not be nil")

	// unparsible SkipVerify
	params[SkipVerify] = "ttt"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// parsible SkipVerify
	params[SkipVerify] = "true"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite shouldn't be nil")

	// unparsible persistOnError
	params[PersistOnError] = "ttt"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// parsible persistOnError
	params[PersistOnError] = "true"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite shouldn't be nil")

	// unparsible StoreEventTags
	params[StoreEventTags] = "ttt"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// parsible StoreEventTags
	params[StoreEventTags] = "true"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite shouldn't be nil")

	// unparsible StoreReadingTags
	params[StoreReadingTags] = "ttt"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.Nil(t, trx, "return result from InfluxDBSyncWrite should be nil")

	// parsible StoreReadingTags
	params[StoreReadingTags] = "true"
	trx = configurable.InfluxDBSyncWrite(params)
	assert.NotNil(t, trx, "return result from InfluxDBSyncWrite shouldn't be nil")
}
