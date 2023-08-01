// Copyright (C) 2020-2021 IOTech Ltd

package xpert

import (
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	floatValue         = "3.999"
	expectedIntValue   = "3"
	intValue           = "123"
	expectedFloatValue = "1.230000e+02"
	uintValue          = "123"
	boolValueTrue      = "true"
	boolValueFalse     = "false"
	stringValue        = "abcd1234"
	devID1             = "id1"
)

func TestConvertToAWSDeviceShadow(t *testing.T) {
	tests := []struct {
		Name             string
		params           interface{}
		ErrorExpectation bool
	}{
		{"No params", nil, true},
		{"Wrong input", "data", true},
		{"Right input", dtos.Event{}, false},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			continuePipeline, result := NewConversion().ConvertToAWSDeviceShadow(ctx, test.params)
			if test.ErrorExpectation {
				assert.Equal(t, continuePipeline, false)
			} else {
				assert.Equal(t, continuePipeline, true)
				assert.IsType(t, []byte{}, result)
				var content map[string]map[string]interface{}
				err := json.Unmarshal(result.([]byte), &content)
				assert.NoError(t, err)
				state, ok := content["state"]
				assert.True(t, ok)
				_, ok = state["reported"]
				assert.True(t, ok)
			}
		})
	}
}

func TestConvertBoolToIntReadingBool(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeBool, SimpleReading: dtos.SimpleReading{Value: boolValueTrue}},
		dtos.BaseReading{ValueType: common.ValueTypeBool, SimpleReading: dtos.SimpleReading{Value: boolValueFalse}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)
	assert.Equal(t, common.ValueTypeInt8, result.(dtos.Event).Readings[0].ValueType)
	assert.Equal(t, BoolTrueNumericValue, result.(dtos.Event).Readings[0].Value)
	assert.Equal(t, common.ValueTypeInt8, result.(dtos.Event).Readings[1].ValueType)
	assert.Equal(t, BoolFalseNumericValue, result.(dtos.Event).Readings[1].Value)
}

func TestConvertBoolToIntReadingInt(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeInt8, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt16, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt32, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt64, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint8, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint16, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint32, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint64, SimpleReading: dtos.SimpleReading{Value: intValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, intValue, reading.Value)
	}
}

func TestConvertBoolToIntReadingFloat(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: floatValue}},
		dtos.BaseReading{ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: floatValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, floatValue, reading.Value)
	}
}

func TestConvertBoolToIntReadingString(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeString, SimpleReading: dtos.SimpleReading{Value: stringValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, stringValue, reading.Value)
	}
}

func TestConvertBoolToIntReadingNoParams(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToIntReading(ctx, nil)

	require.False(t, continuePipeline)
	assert.Equal(t, "no Event Received", result.(error).Error())
}

func TestConvertBoolToIntReadingNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToIntReading(ctx, "")

	require.False(t, continuePipeline)
	assert.Equal(t, "type received is not an Event", result.(error).Error())
}

func TestConvertBoolToFloatReadingBool(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeBool, SimpleReading: dtos.SimpleReading{Value: boolValueTrue}},
		dtos.BaseReading{ValueType: common.ValueTypeBool, SimpleReading: dtos.SimpleReading{Value: boolValueFalse}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)
	assert.Equal(t, common.ValueTypeFloat32, result.(dtos.Event).Readings[0].ValueType)
	assert.Equal(t, BoolTrueNumericValue, result.(dtos.Event).Readings[0].Value)
	assert.Equal(t, common.ValueTypeFloat32, result.(dtos.Event).Readings[1].ValueType)
	assert.Equal(t, BoolFalseNumericValue, result.(dtos.Event).Readings[1].Value)
}

func TestConvertBoolToFloatReadingInt(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeInt8, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt16, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt32, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt64, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint8, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint16, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint32, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint64, SimpleReading: dtos.SimpleReading{Value: intValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, intValue, reading.Value)
	}
}

func TestConvertBoolToFloatReadingFloat(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: floatValue}},
		dtos.BaseReading{ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: floatValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, floatValue, reading.Value)
	}
}

func TestConvertBoolToFloatReadingString(t *testing.T) {

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeString, SimpleReading: dtos.SimpleReading{Value: stringValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, stringValue, reading.Value)
	}
}

func TestConvertBoolToFloatReadingNoParams(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToFloatReading(ctx, nil)

	require.False(t, continuePipeline)
	assert.Equal(t, "no Event Received", result.(error).Error())
}

func TestConvertBoolToFloatReadingNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertBoolToFloatReading(ctx, "")

	require.False(t, continuePipeline)
	assert.Equal(t, "type received is not an Event", result.(error).Error())

}

func TestConvertIntToFloatReadingInt(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeInt8, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt16, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt32, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt64, SimpleReading: dtos.SimpleReading{Value: intValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertIntToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)
	for _, r := range result.(dtos.Event).Readings {
		assert.Equal(t, common.ValueTypeFloat64, r.ValueType)
		assert.Equal(t, expectedFloatValue, r.Value)
	}
}

func TestConvertIntToFloatReadingUint(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeUint8, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint16, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint32, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint64, SimpleReading: dtos.SimpleReading{Value: intValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertIntToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, r := range result.(dtos.Event).Readings {
		assert.Equal(t, common.ValueTypeFloat64, r.ValueType)
		assert.Equal(t, expectedFloatValue, r.Value)
	}
}

func TestConvertIntToFloatReadingFloat(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: floatValue}},
		dtos.BaseReading{ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: floatValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertIntToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, floatValue, reading.Value)
	}
}

func TestConvertIntToFloatReadingString(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeString, SimpleReading: dtos.SimpleReading{Value: stringValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertIntToFloatReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, stringValue, reading.Value)
	}
}

func TestConvertIntToFloatReadingNoParams(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertIntToFloatReading(ctx, nil)

	require.False(t, continuePipeline)
	assert.Equal(t, "no Event Received", result.(error).Error())
}

func TestConvertIntToFloatReadingNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertIntToFloatReading(ctx, "")

	require.False(t, continuePipeline)
	assert.Equal(t, "type received is not an Event", result.(error).Error())
}

func TestConvertFloatToIntReadingFloat(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: floatValue}},
		dtos.BaseReading{ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: floatValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)
	for _, r := range result.(dtos.Event).Readings {
		assert.Equal(t, common.ValueTypeInt64, r.ValueType)
		assert.Equal(t, expectedIntValue, r.Value)
	}
}

func TestConvertFloatToIntReadingExceedsMaxInt(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	testValue := "9.223373e+18"
	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: testValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, eventIn)

	require.False(t, continuePipeline)
	assert.Contains(t, result.(error).Error(), "outside the range of Int64")
}

func TestConvertFloatToIntReadingExceedsMinInt(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	testValue := "-9.223373e+18"
	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: testValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, eventIn)

	require.False(t, continuePipeline)
	assert.Contains(t, result.(error).Error(), "outside the range of Int64")
}

func TestConvertFloatToIntReadingInt(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeInt8, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt16, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt32, SimpleReading: dtos.SimpleReading{Value: intValue}},
		dtos.BaseReading{ValueType: common.ValueTypeInt64, SimpleReading: dtos.SimpleReading{Value: intValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)
	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, intValue, reading.Value)
	}
}

func TestConvertFloatToIntReadingUint(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeUint8, SimpleReading: dtos.SimpleReading{Value: uintValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint16, SimpleReading: dtos.SimpleReading{Value: uintValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint32, SimpleReading: dtos.SimpleReading{Value: uintValue}},
		dtos.BaseReading{ValueType: common.ValueTypeUint64, SimpleReading: dtos.SimpleReading{Value: uintValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, uintValue, reading.Value)
	}
}

func TestConvertFloatToIntReadingString(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeString, SimpleReading: dtos.SimpleReading{Value: stringValue}},
	)
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, eventIn)

	require.True(t, continuePipeline)
	assert.NotNil(t, result)

	for _, reading := range result.(dtos.Event).Readings {
		assert.Equal(t, stringValue, reading.Value)
	}
}

func TestConvertFloatToIntReadingNoParams(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, nil)

	require.False(t, continuePipeline)
	assert.Equal(t, "no Event Received", result.(error).Error())
}

func TestConvertFloatToIntReadingNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.ConvertFloatToIntReading(ctx, "")

	require.False(t, continuePipeline)
	assert.Equal(t, "type received is not an Event", result.(error).Error())
}

func TestConvertByteArrayToEvent(t *testing.T) {
	var emptyBytes []byte
	var eventBytes []byte
	var err error
	eventBytes, err = json.Marshal(dtos.Event{})
	require.NoError(t, err)

	conv := NewConversion()
	tests := []struct {
		Name             string
		Data             interface{}
		ErrorExpectation bool
	}{
		{"no data receive", nil, true},
		{"type received is not supported", "", true},
		{"not an event", emptyBytes, true},
		{"successful case", eventBytes, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			continuePipeline, result := conv.ConvertByteArrayToEvent(ctx, test.Data)
			if test.ErrorExpectation {
				assert.Equal(t, continuePipeline, false)
			} else {
				assert.Equal(t, continuePipeline, true)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestScriptTransform(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}
	eventIn.Readings = append(eventIn.Readings,
		dtos.BaseReading{ValueType: common.ValueTypeUint8, SimpleReading: dtos.SimpleReading{Value: intValue}},
	)
	successfulTransformScript := "var outputObject = { value: inputObject.readings[0].value}; return outputObject;"
	expectedOutputString := "{\"value\":\"" + intValue + "\"}"

	failedTransformScript := "var outputObject = { value: inputObject.readings[3].value}; return outputObject;"

	tests := []struct {
		Name                 string
		Data                 dtos.Event
		Script               string
		ExpectedOutputString string
		ErrorExpectation     bool
	}{
		{"successful case", eventIn, successfulTransformScript, expectedOutputString, false},
		{"failed case - properties of undefined", eventIn, failedTransformScript, "", true},
		{"failed case - empty script", eventIn, "", "", true},
		{"failed case - undefined", eventIn, "aaa", "", true},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			conv := NewScriptTransform(test.Script)
			continuePipeline, result := conv.Transform(ctx, test.Data)

			if test.ErrorExpectation {
				assert.Equal(t, continuePipeline, false)
			} else {
				assert.Equal(t, continuePipeline, true)
				assert.NotNil(t, result)
				assert.Equal(t, result, []byte(test.ExpectedOutputString))
			}
		})
	}
}
