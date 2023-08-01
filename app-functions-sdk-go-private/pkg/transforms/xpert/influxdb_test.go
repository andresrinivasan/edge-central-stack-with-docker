// Copyright (C) 2021-2023 IOTech Ltd

package xpert

import (
	"fmt"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testProfileName       = "testProfile"
	testDeviceName        = "testDevice"
	testResourceName      = "testResource"
	testSecretPath        = "influxdb"
	testMeasurement       = "measurement"
	testInfluxDBAuthToken = "myuser:mypassword"
	// iotech logo encoded with base64
	binaryFileBase64Encoded = "iVBORw0KGgoAAAANSUhEUgAAAXoAAACECAYAAACeT+BNAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyhpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNi1jMTM4IDc5LjE1OTgyNCwgMjAxNi8wOS8xNC0wMTowOTowMSAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIDIwMTcgKE1hY2ludG9zaCkiIHhtcE1NOkluc3RhbmNlSUQ9InhtcC5paWQ6Qjk0QzNDNkYxNjUyMTFFNzlERDNEODRCNTgwMEZENjkiIHhtcE1NOkRvY3VtZW50SUQ9InhtcC5kaWQ6Qjk0QzNDNzAxNjUyMTFFNzlERDNEODRCNTgwMEZENjkiPiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDpCOTRDM0M2RDE2NTIxMUU3OUREM0Q4NEI1ODAwRkQ2OSIgc3RSZWY6ZG9jdW1lbnRJRD0ieG1wLmRpZDpCOTRDM0M2RTE2NTIxMUU3OUREM0Q4NEI1ODAwRkQ2OSIvPiA8L3JkZjpEZXNjcmlwdGlvbj4gPC9yZGY6UkRGPiA8L3g6eG1wbWV0YT4gPD94cGFja2V0IGVuZD0iciI/PmYRBvEAACfFSURBVHja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"
)

func TestInfluxDBSyncWriteNoParams(t *testing.T) {
	writer := NewInfluxDBWriter(InfluxDBWriterConfig{}, true)
	continuePipeline, result := writer.InfluxDBSyncWrite(ctx, nil)
	require.False(t, continuePipeline)
	require.Error(t, result.(error))
}

func TestInfluxDBSyncWriteSetRetryDataPersistFalse(t *testing.T) {
	writer := NewInfluxDBWriter(InfluxDBWriterConfig{}, false)
	ctx.SetRetryData(nil)
	writer.setRetryData(ctx, []byte("data"))
	assert.Nil(t, ctx.RetryData())
}

func TestInfluxDBSyncWriteSetRetryDataPersistTrue(t *testing.T) {
	writer := NewInfluxDBWriter(InfluxDBWriterConfig{}, true)
	ctx.SetRetryData(nil)
	writer.setRetryData(ctx, []byte("data"))
	assert.Equal(t, []byte("data"), ctx.RetryData())
}

func TestInfluxDBSyncWriteValidateSecrets(t *testing.T) {
	writer := NewInfluxDBWriter(InfluxDBWriterConfig{AuthMode: InfluxDBSecretAuthToken}, false)
	tests := []struct {
		Name             string
		secrets          influxDBSecrets
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"No authToken", influxDBSecrets{}, true, "mandatory secret authentication token is empty"},
		{"With authToken", influxDBSecrets{authToken: testInfluxDBAuthToken}, false, ""},
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

func TestInfluxDBSyncWriteValidateOpts(t *testing.T) {
	writer := NewInfluxDBWriter(InfluxDBWriterConfig{AuthMode: InfluxDBSecretAuthToken, SecretPath: testSecretPath, Precision: time.Millisecond, SkipCertVerify: false}, false)
	// setup mock secret client
	expected := map[string]string{
		InfluxDBSecretAuthToken: testInfluxDBAuthToken,
	}
	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("GetSecret", testSecretPath).Return(expected, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})

	err := writer.initializeInfluxDBClient(ctx)
	require.NoError(t, err)
	assert.NotNil(t, writer.client, "writer.client should not be nil")
	assert.Equal(t, time.Millisecond, writer.client.Options().Precision(), "precision should be equal")
	assert.False(t, writer.client.Options().TLSConfig().InsecureSkipVerify, "SkipCertVerify should be false")
}

func TestInfluxDBSyncWriteGetSecrets(t *testing.T) {
	notFoundSecretPath := "notfound"
	writer := NewInfluxDBWriter(InfluxDBWriterConfig{}, false)
	tests := []struct {
		Name            string
		SecretPath      string
		ExpectedSecrets *influxDBSecrets
		ExpectingError  bool
	}{
		{"No Secrets found", "notfound", nil, true},
		{"With Secrets", testSecretPath, &influxDBSecrets{
			authToken: testInfluxDBAuthToken,
		}, false},
	}
	// setup mock secret client
	expected := map[string]string{
		InfluxDBSecretAuthToken: testInfluxDBAuthToken,
	}
	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("GetSecret", notFoundSecretPath).Return(nil, fmt.Errorf("secrets not found"))
	mockSecretProvider.On("GetSecret", testSecretPath).Return(expected, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			writer.config = InfluxDBWriterConfig{
				SecretPath: test.SecretPath,
				AuthMode:   InfluxDBAuthModeToken,
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

func TestParseReadingValue(t *testing.T) {
	tests := []struct {
		Name             string
		reading          dtos.BaseReading
		expected         interface{}
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{
			"zero reading value string",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: ""},
			},
			nil,
			true,
			"zero-length reading value found",
		},
		{
			"int8",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeInt8,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			int64(-123),
			false,
			"",
		},
		{
			"uint8",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeUint8,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			uint64(123),
			false,
			"",
		},
		{
			"int16",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeInt16,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			int64(-123),
			false,
			"",
		},
		{
			"uint16",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeUint16,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			uint64(123),
			false,
			"",
		},
		{
			"int32",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeInt32,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			int64(-123),
			false,
			"",
		},
		{
			"uint32",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeUint8,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			uint64(123),
			false,
			"",
		},
		{
			"int64",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeInt64,
				SimpleReading: dtos.SimpleReading{Value: "-123"},
			},
			int64(-123),
			false,
			"",
		},
		{
			"uint64",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeUint64,
				SimpleReading: dtos.SimpleReading{Value: "123"},
			},
			uint64(123),
			false,
			"",
		},
		{
			"float32 eNotation",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
			},
			float32(-10.600001),
			false,
			"",
		},
		{
			"float64 eNotation",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat64,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
			},
			float64(-10.600001),
			false,
			"",
		},
		{
			"boolean false",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeBool,
				SimpleReading: dtos.SimpleReading{Value: "false"},
			},
			false,
			false,
			"",
		},
		{
			"boolean true",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeBool,
				SimpleReading: dtos.SimpleReading{Value: "true"},
			},
			true,
			false,
			"",
		},
		{
			"boolean unparsable",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeBool,
				SimpleReading: dtos.SimpleReading{Value: "unparsable"},
			},
			nil,
			true,
			"failed to parse reading value",
		},
		{
			"binary",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeBinary,
				BinaryReading: dtos.BinaryReading{BinaryValue: []byte(binaryFileBase64Encoded), MediaType: "File"},
			},
			[]byte(binaryFileBase64Encoded),
			false,
			"",
		},
		{
			"Bool array",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeBoolArray,
				SimpleReading: dtos.SimpleReading{Value: fmt.Sprintf("%v", []bool{true, false, true})},
			},
			fmt.Sprintf("%v", []bool{true, false, true}),
			false,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			value, err := parseReadingValue(test.reading)
			if test.ErrorExpectation {
				assert.Error(t, err, "Result should be an error")
				assert.Contains(t, err.Error(), test.ErrorMessage)
			} else {
				assert.Equal(t, test.expected, value, "Should be equal")
			}
		})
	}
}

func TestParseInteger(t *testing.T) {
	tests := []struct {
		Name             string
		readingValueStr  string
		bitSize          int
		signed           bool
		expected         interface{}
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"zero reading value string", "", 8, true, nil, true, "zero-length reading value found."},
		{"parsing int8 success", "-66", 8, true, int64(-66), false, ""},
		{"parsing int8 fail", "abc", 8, true, nil, true, "failed to parse non-empty reading value"},
		{"parsing int16 success", "-66", 16, true, int64(-66), false, ""},
		{"parsing int16 fail", "abc", 16, true, nil, true, "failed to parse non-empty reading value"},
		{"parsing int32 success", "-66", 32, true, int64(-66), false, ""},
		{"parsing int32 fail", "abc", 32, true, nil, true, "failed to parse non-empty reading value"},
		{"parsing int64 success", "-66", 64, true, int64(-66), false, ""},
		{"parsing int64 fail", "abc", 64, true, nil, true, "failed to parse non-empty reading value"},
		{"parsing uint8 success", "66", 8, false, uint64(66), false, ""},
		{"parsing uint8 fail", "abc", 8, false, nil, true, "failed to parse non-empty reading value"},
		{"parsing uint16 success", "66", 16, false, uint64(66), false, ""},
		{"parsing uint16 fail", "abc", 16, false, nil, true, "failed to parse non-empty reading value"},
		{"parsing uint32 success", "66", 32, false, uint64(66), false, ""},
		{"parsing uint32 fail", "abc", 32, false, nil, true, "failed to parse non-empty reading value"},
		{"parsing uint64 success", "66", 64, false, uint64(66), false, ""},
		{"parsing uint64 fail", "abc", 64, false, nil, true, "failed to parse non-empty reading value"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			value, err := parseInteger(test.readingValueStr, test.bitSize, test.signed)
			if test.ErrorExpectation {
				assert.Error(t, err, "Result should be an error")
				assert.Contains(t, err.Error(), test.ErrorMessage)
			} else {
				assert.Equal(t, test.expected, value, "Should be equal")
			}
		})
	}
}

func TestParseFloat32(t *testing.T) {
	tests := []struct {
		Name             string
		readingValueStr  string
		bitSize          int
		expected         interface{}
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"float32 zero reading value string", "", 32, nil, true, "zero-length reading value found"},
		{"float32 enotation parsable", "-1.0600001e1", 32, float32(-10.600001), false, ""},
		{"float32 enotation unparsable", "abc", 32, float32(-10.600001), true, "failed to parse reading value"},
		{"float64 zero reading value string", "", 64, nil, true, "zero-length reading value found"},
		{"float64 enotation parsable", "-1.0600001e1", 64, float64(-10.600001), false, ""},
		{"float64 enotation unparsable", "abc", 64, float64(-10.600001), true, "failed to parse reading value"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			value, err := parseFloat(test.readingValueStr, test.bitSize)
			if test.ErrorExpectation {
				assert.Error(t, err, "Result should be an error")
				assert.Contains(t, err.Error(), test.ErrorMessage)
			} else {
				assert.Equal(t, test.expected, value, "Should be equal")
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		Name             string
		readingValueStr  string
		expected         interface{}
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"zero reading value string", "", nil, true, "zero-length reading value found."},
		{"parse true", "true", true, false, ""},
		{"parse false", "false", false, false, ""},
		{"parse unparsable string", "ttt", nil, true, "failed to parse reading value"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			value, err := parseBool(test.readingValueStr)
			if test.ErrorExpectation {
				assert.Error(t, err, "Result should be an error")
				assert.Contains(t, err.Error(), test.ErrorMessage)
			} else {
				assert.Equal(t, test.expected, value, "Should be equal")
			}
		})
	}
}

func TestToNanoseconds(t *testing.T) {
	tests := []struct {
		Name      string
		timestamp int64
		expected  time.Time
	}{
		{"conversion case 1", int64(1581420776817), time.Unix(0, int64(1581420776817000000))},
		{"conversion case 2", int64(123456789), time.Unix(0, int64(1234567890000000000))},
		{"conversion case 3", int64(12345678912345), time.Unix(0, int64(1234567891234500000))},
		{"conversion case 4", int64(1), time.Unix(0, int64(1000000000000000000))},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := toNanosecondsTime(test.timestamp)
			assert.True(t, result.Equal(test.expected), "Should be equal")
		})
	}
}

func TestIsAcceptedReadingValueType(t *testing.T) {
	tests := []struct {
		Name              string
		readingValueType  string
		influxDBValueType string
		expectedResult    bool
	}{
		{"float true", common.ValueTypeFloat32, InfluxDBValueTypeFloat, true},
		{"float false", common.ValueTypeFloat32Array, InfluxDBValueTypeFloat, false},
		{"int8 true", common.ValueTypeInt8, InfluxDBValueTypeInteger, true},
		{"int16 true", common.ValueTypeInt16, InfluxDBValueTypeInteger, true},
		{"int32 true", common.ValueTypeInt32, InfluxDBValueTypeInteger, true},
		{"int64 true", common.ValueTypeInt64, InfluxDBValueTypeInteger, true},
		{"uint8 true", common.ValueTypeUint8, InfluxDBValueTypeUInteger, true},
		{"uint16 true", common.ValueTypeUint16, InfluxDBValueTypeUInteger, true},
		{"uint32 true", common.ValueTypeUint32, InfluxDBValueTypeUInteger, true},
		{"uint64 true", common.ValueTypeUint64, InfluxDBValueTypeUInteger, true},
		{"int false", common.ValueTypeInt32Array, InfluxDBValueTypeInteger, false},
		{"string true", common.ValueTypeString, InfluxDBValueTypeString, true},
		{"int32 array string true", common.ValueTypeInt32Array, InfluxDBValueTypeString, true},
		{"string false", common.ValueTypeInt32, InfluxDBValueTypeString, false},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := isAcceptedReadingValueType(test.readingValueType, test.influxDBValueType)
			if test.expectedResult {
				assert.True(t, result, "Result should be true")
			} else {
				assert.False(t, result, "Result should be false")
			}
		})
	}
}

func TestToPoint(t *testing.T) {
	eventTagKeys := []string{"t1", "t2"}
	eventTags := map[string]interface{}{
		eventTagKeys[0]: 123,
		eventTagKeys[1]: "tt",
	}
	readingTagKeys := []string{"r1", "r2"}
	readingTags := map[string]interface{}{
		readingTagKeys[0]: 456,
		readingTagKeys[1]: "rr",
	}

	tests := []struct {
		Name             string
		reading          dtos.BaseReading
		fieldKeyPattern  string
		storeEventTags   bool
		storeReadingTags bool
		eventTags        map[string]interface{}
	}{
		{
			"storeEventTags is true",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
			},
			"value",
			true,
			false,
			eventTags,
		},
		{
			"storeEventTags is false",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
			},
			"value",
			false,
			false,
			eventTags,
		},
		{
			"storeReadingTags is true",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
				Tags:          readingTags,
			},
			"value",
			false,
			true,
			eventTags,
		},
		{
			"storeReadingTags and storeEventTags are true",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
				Tags:          readingTags,
			},
			"value",
			true,
			true,
			eventTags,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			point, err := toPoint(test.reading, testMeasurement, test.fieldKeyPattern, test.storeEventTags,
				test.eventTags, test.storeReadingTags)
			require.NoError(t, err, "requires no error returned")

			tagList := []string{}
			for _, tag := range point.TagList() {
				tagList = append(tagList, tag.Key)
			}

			if test.storeReadingTags {
				assert.Contains(t, tagList, readingTagKeys[0])
				assert.Contains(t, tagList, readingTagKeys[1])
			} else {
				assert.NotContains(t, tagList, readingTagKeys[0])
				assert.NotContains(t, tagList, readingTagKeys[1])
			}

			if test.storeEventTags {
				assert.Contains(t, tagList, eventTagKeys[0])
				assert.Contains(t, tagList, eventTagKeys[1])
			} else {
				assert.NotContains(t, tagList, eventTagKeys[0])
				assert.NotContains(t, tagList, eventTagKeys[1])
			}
		})
	}
}

func TestToPointFieldKey(t *testing.T) {
	tests := []struct {
		Name             string
		reading          dtos.BaseReading
		fieldKeyPattern  string
		expectedFieldKey string
	}{
		{
			"fieldKeyPattern specifies fixed value",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
			},
			"value",
			"value",
		},
		{
			"fieldKeyPattern specifies replaceable variables",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
			},
			"{resourceName}{deviceName}{profileName}{valueType}",
			testResourceName + testDeviceName + testProfileName + common.ValueTypeFloat32,
		},
		{
			"fieldKeyPattern specifies double replaceable variables",
			dtos.BaseReading{
				ProfileName:   testProfileName,
				DeviceName:    testDeviceName,
				ResourceName:  testResourceName,
				ValueType:     common.ValueTypeFloat32,
				SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"},
			},
			"{resourceName}{deviceName}{profileName}{valueType}{valueType}",
			testResourceName + testDeviceName + testProfileName + common.ValueTypeFloat32 + common.ValueTypeFloat32,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fieldKey := toPointFieldKey(test.reading, test.fieldKeyPattern)
			assert.Equal(t, test.expectedFieldKey, fieldKey)
		})
	}
}
