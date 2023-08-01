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
	IntMIN     = "-10"
	IntMAX     = "50"
	UintMIN    = "50"
	UintMAX    = "100"
	Float32MIN = "-10.6"
	Float32MAX = "500.89"
	Float64MIN = "-1000.456"
	Float64MAX = "3000.789"
)

func TestFilterByValueMaxMinInt(t *testing.T) {

	expected := dtos.Event{
		DeviceName: devID1,
	}

	expected.Readings = append(expected.Readings,
		dtos.BaseReading{ResourceName: common.ValueTypeInt8, ValueType: common.ValueTypeInt8, SimpleReading: dtos.SimpleReading{Value: "-47"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeInt8, ValueType: common.ValueTypeInt8, SimpleReading: dtos.SimpleReading{Value: "47"}},
		dtos.BaseReading{ResourceName: common.ValueTypeInt16, ValueType: common.ValueTypeInt16, SimpleReading: dtos.SimpleReading{Value: "-5"}},
		dtos.BaseReading{ResourceName: common.ValueTypeInt16, ValueType: common.ValueTypeInt16, SimpleReading: dtos.SimpleReading{Value: "99"}},  //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeInt32, ValueType: common.ValueTypeInt32, SimpleReading: dtos.SimpleReading{Value: "-55"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeInt32, ValueType: common.ValueTypeInt32, SimpleReading: dtos.SimpleReading{Value: "100"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeInt64, ValueType: common.ValueTypeInt64, SimpleReading: dtos.SimpleReading{Value: "-1"}},
		dtos.BaseReading{ResourceName: common.ValueTypeInt64, ValueType: common.ValueTypeInt64, SimpleReading: dtos.SimpleReading{Value: "49"}},
	)

	filter := NewFilter()

	continuePipeline, result := filter.FilterByValueMaxMin(ctx, expected)
	require.NotNil(t, result, "Expected event to be passed thru")

	actual, ok := result.(dtos.Event)
	require.True(t, ok, "Expected result to be an Event")
	require.NotNil(t, actual.Readings, "Expected Reading passed thru")
	assert.Equal(t, expected.DeviceName, actual.DeviceName, "Expected Event to be same as passed in")
	assert.Equal(t, 4, len(actual.Readings), "Expected 4 readings (outlier) will be filtered out")
	assert.True(t, continuePipeline, "Pipeline should'nt stop processing")
}

func TestFilterByValueMaxMinUint(t *testing.T) {

	expected := dtos.Event{
		DeviceName: devID1,
	}

	expected.Readings = append(expected.Readings,
		dtos.BaseReading{ResourceName: common.ValueTypeUint8, ValueType: common.ValueTypeUint8, SimpleReading: dtos.SimpleReading{Value: "49"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeUint8, ValueType: common.ValueTypeUint8, SimpleReading: dtos.SimpleReading{Value: "77"}},
		dtos.BaseReading{ResourceName: common.ValueTypeUint16, ValueType: common.ValueTypeUint16, SimpleReading: dtos.SimpleReading{Value: "99"}},
		dtos.BaseReading{ResourceName: common.ValueTypeUint16, ValueType: common.ValueTypeUint16, SimpleReading: dtos.SimpleReading{Value: "101"}},  //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeUint32, ValueType: common.ValueTypeUint32, SimpleReading: dtos.SimpleReading{Value: "1"}},    //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeUint32, ValueType: common.ValueTypeUint32, SimpleReading: dtos.SimpleReading{Value: "1000"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeUint64, ValueType: common.ValueTypeUint64, SimpleReading: dtos.SimpleReading{Value: "88"}},
		dtos.BaseReading{ResourceName: common.ValueTypeUint64, ValueType: common.ValueTypeUint64, SimpleReading: dtos.SimpleReading{Value: "91"}},
	)

	filter := NewFilter()

	continuePipeline, result := filter.FilterByValueMaxMin(ctx, expected)
	require.NotNil(t, result, "Expected event to be passed thru")

	actual, ok := result.(dtos.Event)
	require.True(t, ok, "Expected result to be an Event")
	require.NotNil(t, actual.Readings, "Expected Reading passed thru")
	assert.Equal(t, expected.DeviceName, actual.DeviceName, "Expected Event to be same as passed in")
	assert.Equal(t, 4, len(actual.Readings), "Expected 4 readings (outlier) will be filtered out")
	assert.True(t, continuePipeline, "Pipeline should'nt stop processing")
}

func TestFilterByValueMaxMinFloat(t *testing.T) {

	expected := dtos.Event{
		DeviceName: devID1,
	}

	expected.Readings = append(expected.Readings,
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "-10.600001"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "-10.599999"}},
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "500.888889"}},
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "500.891"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "-1000.4559999"}},
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "-1000.4560001"}}, //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "3000.7888889"}},
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "3000.7890001"}}, //outlier
	)

	filter := NewFilter()

	continuePipeline, result := filter.FilterByValueMaxMin(ctx, expected)
	require.NotNil(t, result, "Expected event to be passed thru")

	actual, ok := result.(dtos.Event)
	require.True(t, ok, "Expected result to be an Event")
	require.NotNil(t, actual.Readings, "Expected Reading passed thru")
	assert.Equal(t, expected.DeviceName, actual.DeviceName, "Expected Event to be same as passed in")
	assert.Equal(t, 4, len(actual.Readings), "Expected 4 readings (outlier) will be filtered out")
	assert.True(t, continuePipeline, "Pipeline should'nt stop processing")
}

func TestFilterByValueMaxMinFloatENotation(t *testing.T) {

	expected := dtos.Event{
		DeviceName: devID1,
	}

	expected.Readings = append(expected.Readings,
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "-1.0600001e1"}},    //-10.600001 outlier
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "-1.0599999e1"}},    //-10.599999
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "5.00888889e2"}},    //500.888889
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "5.00891e2"}},       //500.891 outlier
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "-1.0004559999e3"}}, //-1000.4559999
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "-1.0004560001e3"}}, //-1000.4560001 outlier
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "3.0007888889e3"}},  //3000.7888889
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "3.0007890001e3"}},  //3000.7890001 outlier
	)

	filter := NewFilter()

	continuePipeline, result := filter.FilterByValueMaxMin(ctx, expected)
	require.NotNil(t, result, "Expected event to be passed thru")

	actual, ok := result.(dtos.Event)
	require.True(t, ok, "Expected result to be an Event")
	require.NotNil(t, actual.Readings, "Expected Reading passed thru")
	assert.Equal(t, expected.DeviceName, actual.DeviceName, "Expected Event to be same as passed in")
	assert.Equal(t, 4, len(actual.Readings), "Expected 4 readings (outlier) will be filtered out")
	assert.True(t, continuePipeline, "Pipeline should'nt stop processing")
}

func TestFilterByValueMaxMinNotEvent(t *testing.T) {
	expected := dtos.Event{
		DeviceName: devID1,
	}

	expected.Readings = append(expected.Readings,
		dtos.BaseReading{ResourceName: common.ValueTypeFloat32, ValueType: common.ValueTypeFloat32, SimpleReading: dtos.SimpleReading{Value: "-10.600001"}},   //outlier
		dtos.BaseReading{ResourceName: common.ValueTypeFloat64, ValueType: common.ValueTypeFloat64, SimpleReading: dtos.SimpleReading{Value: "3000.7890001"}}, //outlier
	)

	eventString, _ := json.Marshal(expected)

	filter := NewFilter()

	continuePipeline, result := filter.FilterByValueMaxMin(ctx, eventString)
	_, ok := result.(error)
	require.True(t, ok, "Expected result to be an error")
	assert.False(t, continuePipeline, "Pipeline should stop")
}

func TestFilterByValueMaxMinNoParams(t *testing.T) {
	filter := NewFilter()

	continuePipeline, result := filter.FilterByValueMaxMin(ctx, nil)
	_, ok := result.(error)
	require.True(t, ok, "Expected result to be an error")
	assert.False(t, continuePipeline, "Pipeline should stop")
}
