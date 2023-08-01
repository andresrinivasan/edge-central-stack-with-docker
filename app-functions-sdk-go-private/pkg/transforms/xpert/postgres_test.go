// Copyright (C) 2022-2023 IOTech Ltd

package xpert

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	sdkMocks "github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostgresWriteNoParams(t *testing.T) {
	writer := NewPostgresWriter(PostgresWriterConfig{}, true)
	continuePipeline, result := writer.PostgresWrite(ctx, nil)
	require.False(t, continuePipeline)
	require.Error(t, result.(error))
}

func TestPostgresWriteSetRetryDataPersistFalse(t *testing.T) {
	writer := NewPostgresWriter(PostgresWriterConfig{}, false)
	ctx.SetRetryData(nil)
	writer.setRetryData(ctx, []byte("data"))
	require.Nil(t, ctx.RetryData())
}

func TestPostgresWriteSetRetryDataPersistTrue(t *testing.T) {
	writer := NewPostgresWriter(PostgresWriterConfig{}, true)
	ctx.SetRetryData(nil)
	writer.setRetryData(ctx, []byte("data"))
	require.Equal(t, []byte("data"), ctx.RetryData())
}

func TestPostgresWriteValidateSecrets(t *testing.T) {
	writer := NewPostgresWriter(PostgresWriterConfig{AuthMode: PostgresAuthModeUsernamePassword}, false)
	tests := []struct {
		Name             string
		secrets          postgresSecrets
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"No username", postgresSecrets{password: "password"}, true,
			fmt.Sprintf("auth mode %s selected however %s was not found at secret path",
				PostgresAuthModeUsernamePassword, messaging.SecretUsernameKey)},
		{"No password", postgresSecrets{username: "username"}, true,
			fmt.Sprintf("auth mode %s selected however %s was not found at secret path",
				PostgresAuthModeUsernamePassword, messaging.SecretPasswordKey)},
		{"With username and password", postgresSecrets{username: "username", password: "password"}, false, ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := writer.validateSecrets(test.secrets)
			if test.ErrorExpectation {
				assert.Error(t, result, "Result should be an error")
				assert.Equal(t, test.ErrorMessage, result.Error())
			} else {
				assert.Nil(t, result, "Should be nil")
			}
		})
	}
}

func TestPostgresWriteGetSecrets(t *testing.T) {
	secretPath := "postgres"
	notFoundSecretPath := "notfound"
	username := "username"
	password := "password"
	writer := NewPostgresWriter(PostgresWriterConfig{}, false)
	tests := []struct {
		Name            string
		SecretPath      string
		ExpectedSecrets *postgresSecrets
		ExpectingError  bool
	}{
		{"No Secrets found", "notfound", nil, true},
		{"With Secrets", secretPath, &postgresSecrets{
			username: username,
			password: password,
		}, false},
	}
	// setup mock secret client
	expected := map[string]string{
		messaging.SecretUsernameKey: username,
		messaging.SecretPasswordKey: password,
	}
	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("GetSecret", notFoundSecretPath).Return(nil, fmt.Errorf("secrets not found"))
	mockSecretProvider.On("GetSecret", secretPath).Return(expected, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			writer.config = PostgresWriterConfig{
				SecretPath: test.SecretPath,
				AuthMode:   messaging.AuthModeUsernamePassword,
			}
			secrets, err := writer.getSecrets(ctx)
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			require.Equal(t, test.ExpectedSecrets, secrets)
		})
	}
}

func TestGetSqlStatementAndArguments(t *testing.T) {
	eventId := "18138082-d85c-49bb-96fa-4204cffc9307"
	deviceName := "testDevice"
	configStoreEventIdFalse := PostgresWriterConfig{StoreEventId: false}
	configStoreEventIdTrue := PostgresWriterConfig{StoreEventId: true}
	tags := map[string]interface{}{
		"country": "UK",
		"rank":    1,
	}

	testFloatValue := -1.0600001e1
	// float64 test data
	var testFloat64BinaryValue [8]byte
	binary.BigEndian.PutUint64(testFloat64BinaryValue[:], math.Float64bits(testFloatValue))
	// float32 test data
	var testFloat32BinaryValue [4]byte
	binary.BigEndian.PutUint32(testFloat32BinaryValue[:], math.Float32bits(float32(testFloatValue)))

	testObjectValue := map[string]interface{}{
		"attr1": "value1",
		"attr2": -45,
		"attr3": []interface{}{255, 1, 0},
	}
	testObjectBinaryValue, err := json.Marshal(testObjectValue)
	assert.NoError(t, err)
	testObjectArrayBinaryValue, err := json.Marshal([]map[string]interface{}{testObjectValue, testObjectValue})
	assert.NoError(t, err)
	tests := []struct {
		Name          string
		config        PostgresWriterConfig
		reading       dtos.BaseReading
		expectedValue []byte
	}{
		{
			"boolean",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeBool,
				ValueType:     common.ValueTypeBool,
				SimpleReading: dtos.SimpleReading{Value: "false"},
			},
			[]byte{0},
		},
		{
			"Bool array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeBoolArray,
				ValueType:     common.ValueTypeBoolArray,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]bool{true, false, true})},
			},
			[]byte(arrayToString([]bool{true, false, true})),
		},
		{
			"string",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeString,
				ValueType:     common.ValueTypeString,
				SimpleReading: dtos.SimpleReading{Value: "string"},
			},
			[]byte("string"),
		},
		{
			"string array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeStringArray,
				ValueType:     common.ValueTypeStringArray,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]string{"abc", "123"})},
			},
			[]byte(arrayToString([]string{"abc", "123"})),
		},
		{
			"int8",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt8,
				ValueType:     common.ValueTypeInt8,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			[]byte{133},
		},
		{
			"int8 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt8Array,
				ValueType:     common.ValueTypeInt8Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]int8{1, 23})},
			},
			[]byte(arrayToString([]int8{1, 23})),
		},
		{
			"uint8",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint8,
				ValueType:     common.ValueTypeUint8,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			[]byte{123},
		},
		{
			"uint8 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint8Array,
				ValueType:     common.ValueTypeUint8Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]uint8{1, 23})},
			},
			[]byte(arrayToString([]uint8{1, 23})),
		},
		{
			"int16",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt16,
				ValueType:     common.ValueTypeInt16,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			[]byte{255, 133},
		},
		{
			"int16 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt16Array,
				ValueType:     common.ValueTypeInt16Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]int16{1, 23})},
			},
			[]byte(arrayToString([]int16{1, 23})),
		},
		{
			"uint16",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint16,
				ValueType:     common.ValueTypeUint16,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			[]byte{0, 123},
		},
		{
			"uint16 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint16Array,
				ValueType:     common.ValueTypeUint16Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]uint16{1, 23})},
			},
			[]byte(arrayToString([]uint16{1, 23})),
		},
		{
			"int32",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt32,
				ValueType:     common.ValueTypeInt32,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			[]byte{255, 255, 255, 133},
		},
		{
			"int32 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt32Array,
				ValueType:     common.ValueTypeInt32Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]int32{1, 23})},
			},
			[]byte(arrayToString([]int32{1, 23})),
		},
		{
			"uint32",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint32,
				ValueType:     common.ValueTypeUint32,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			[]byte{0, 0, 0, 123},
		},
		{
			"uint32 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint32Array,
				ValueType:     common.ValueTypeUint32Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]uint32{1, 23})},
			},
			[]byte(arrayToString([]uint32{1, 23})),
		},
		{
			"int64",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt64,
				ValueType:     common.ValueTypeInt64,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			[]byte{255, 255, 255, 255, 255, 255, 255, 133},
		},
		{
			"int64 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeInt64Array,
				ValueType:     common.ValueTypeInt64Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]int64{1, 23})},
			},
			[]byte(arrayToString([]int64{1, 23})),
		},
		{
			"uint64",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint64,
				ValueType:     common.ValueTypeUint64,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			[]byte{0, 0, 0, 0, 0, 0, 0, 123},
		},
		{
			"uint64 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeUint64Array,
				ValueType:     common.ValueTypeUint64Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]uint64{1, 23})},
			},
			[]byte(arrayToString([]uint64{1, 23})),
		},
		{
			"float32",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeFloat32,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: fmt.Sprintf("%f", testFloatValue)},
			},
			testFloat32BinaryValue[:],
		},
		{
			"float32 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeFloat32Array,
				ValueType:     common.ValueTypeFloat32Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]float32{float32(testFloatValue), float32(testFloatValue)})},
			},
			[]byte(arrayToString([]float32{float32(testFloatValue), float32(testFloatValue)})),
		},
		{
			"float64",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeFloat64,
				ValueType:     common.ValueTypeFloat64,
				SimpleReading: dtos.SimpleReading{Value: fmt.Sprintf("%f", testFloatValue)},
			},
			testFloat64BinaryValue[:],
		},
		{
			"float64 array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeFloat64Array,
				ValueType:     common.ValueTypeFloat64Array,
				SimpleReading: dtos.SimpleReading{Value: arrayToString([]float64{testFloatValue, testFloatValue})},
			},
			[]byte(arrayToString([]float64{testFloatValue, testFloatValue})),
		},
		{
			"binary",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeBinary,
				ValueType:     common.ValueTypeBinary,
				BinaryReading: dtos.BinaryReading{BinaryValue: []byte("binary")},
			},
			[]byte("binary"),
		},
		{
			"object",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeObject,
				ValueType:     common.ValueTypeObject,
				ObjectReading: dtos.ObjectReading{ObjectValue: testObjectValue},
			},
			testObjectBinaryValue,
		},
		{
			"object array",
			configStoreEventIdFalse,
			dtos.BaseReading{
				DeviceName:   deviceName,
				ResourceName: common.ValueTypeObjectArray,
				ValueType:    common.ValueTypeObjectArray,
				ObjectReading: dtos.ObjectReading{ObjectValue: []map[string]interface{}{
					testObjectValue, testObjectValue,
				}},
			},
			testObjectArrayBinaryValue,
		},
		{
			"store event id",
			configStoreEventIdTrue,
			dtos.BaseReading{
				DeviceName:    deviceName,
				ResourceName:  common.ValueTypeString,
				ValueType:     common.ValueTypeString,
				SimpleReading: dtos.SimpleReading{Value: "string"},
			},
			[]byte("string"),
		},
	}

	mockCacheHelper := &sdkMocks.PostgresCacheHelper{}
	mockResourceId := int32(1)
	mockValueTypeId := int16(1)
	mockCacheHelper.On("GetResourceId", mock.Anything, mock.Anything).Return(mockResourceId, nil)
	mockCacheHelper.On("GetValueTypeId", mock.Anything).Return(mockValueTypeId, nil)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			sql, args, err := getSqlStatementAndArguments(mockCacheHelper, test.config, test.reading, eventId, tags)
			assert.NoError(t, err)
			if test.config.StoreEventId {
				assert.Contains(t, sql, PostgresColumnEventId)
				assert.Equal(t, eventId, args.eventId)
			} else {
				assert.NotContains(t, sql, PostgresColumnEventId)
				assert.Empty(t, args.eventId)
			}
			assert.Equal(t, test.expectedValue, args.value)
			assert.Equal(t, mockValueTypeId, args.valueTypeId)
			assert.Equal(t, mockResourceId, args.resourceId)
			assert.Equal(t, toNanosecondsTime(test.reading.Origin), args.timestamp)
		})
	}
}

func arrayToString(value interface{}) string {
	result := fmt.Sprintf("%v", value)
	return strings.ReplaceAll(result, " ", ", ")
}
