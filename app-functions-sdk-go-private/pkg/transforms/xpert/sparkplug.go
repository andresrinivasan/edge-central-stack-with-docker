// Copyright (C) 2022-2023 IOTech Ltd

package xpert

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/sparkplug/protobuf"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"math"
	"reflect"
	"regexp"
)

const (
	Sparkplug                 = "SPARKPLUG"
	SparkplugBTopicLevelDDATA = "DDATA"
)

// SparkplugConverter will transform the incoming Sparkplug message (encoded in protobuf) to edgex event
type SparkplugConverter struct{}

// NewSparkplugConverter creates, initializes and returns a new instance of SparkplugConverter
func NewSparkplugConverter() SparkplugConverter {
	return SparkplugConverter{}
}

func (sc SparkplugConverter) ConvertDDATAtoEvent(ctx interfaces.AppFunctionContext, inputData interface{}) (bool, interface{}) {
	if inputData == nil {
		// We didn't receive a result
		return false, errors.New("no data received")
	}

	bytes, typeAccepted := inputData.([]byte)

	if !typeAccepted {
		return false, fmt.Errorf("unsupported data type passed in: %v.  ConvertDDATAtoEvent expects []byte.", reflect.TypeOf(inputData))
	}

	receivedTopic, found := ctx.GetValue(interfaces.RECEIVEDTOPIC)

	if !found {
		return false, errors.New("received topic not found")
	}

	// topic specification: SparkplugNamespace/GroupID/DDATA/EdgeNodeID/DeviceName
	ok, err := regexp.MatchString(fmt.Sprintf("^([^/\\n]*)/([^/\\n]*)/%s/([^/\\n]*)/([^/\\n]*)$", SparkplugBTopicLevelDDATA), receivedTopic)
	if err != nil {
		return false, fmt.Errorf("fail to check received topic: %s", err.Error())
	}
	if !ok {
		return false, fmt.Errorf("received topic(%s) does not meet the specification of Sparkplug B DDATA topic", receivedTopic)
	}
	r, err := regexp.Compile("([^//]+$)") // find the string after last slash
	if err != nil {
		return false, fmt.Errorf("fail to create a regular expression object:%s", err.Error())
	}
	deviceName := r.FindString(receivedTopic)

	var payload protobuf.Payload
	if err := proto.Unmarshal(bytes, &payload); err != nil {
		return false, fmt.Errorf("fail to unmarshal message to sparkplug protobuf payload.  Error:%s", err.Error())
	}

	event, err := toEvent(deviceName, &payload)

	if err != nil {
		ctx.LoggingClient().Error(err.Error())
		return false, err
	}

	return true, event
}

func toEvent(deviceName string, payload *protobuf.Payload) (dtos.Event, error) {
	event := dtos.Event{
		DeviceName:  deviceName,
		ProfileName: Sparkplug,
		SourceName:  Sparkplug,
		Origin:      int64(*payload.Timestamp),
	}

	if len(payload.Body) > 0 {
		err := json.Unmarshal(payload.Body, &event.Tags)
		if err != nil {
			return event, fmt.Errorf("fail to unmarshal sparkplug payload body as event tags: %s", err.Error())
		}
	}

	for _, metric := range payload.Metrics {
		valueType, value, err := toReadingValueTypeAndValue(metric)
		if err != nil {
			return event, fmt.Errorf("fail to convert the sparkplug metric to a reading value: %s", err.Error())
		}
		if err = event.AddSimpleReading(*metric.Name, valueType, value); err != nil {
			return event, fmt.Errorf("fail to add reading into the event: %s", err.Error())
		}

	}
	if payload.Uuid != nil {
		event.Id = *payload.Uuid
	} else {
		event.Id = uuid.NewString()
	}
	return event, nil
}

// GetMetricValue retrieves metric value in string format from passing metric
func toReadingValueTypeAndValue(metric *protobuf.Payload_Metric) (string, interface{}, error) {
	metricDataType := protobuf.DataType(*metric.Datatype)
	switch metricDataType {
	case protobuf.DataType_Int8:
		var result int64
		if metric.GetIntValue() > math.MaxInt8 {
			// true indicates that the original value is negative, e.g. an int8 -2 is being converted to an uint32
			// 4294967294, which is two integer less than 0.
			result = int64(metric.GetIntValue()) - int64(math.MaxUint32) - 1
			return common.ValueTypeInt8, int8(result), nil
		}
		return common.ValueTypeInt8, int8(metric.GetIntValue()), nil
	case protobuf.DataType_Int16:
		var result int64
		if metric.GetIntValue() > math.MaxInt16 {
			// true indicates that the original value is negative, e.g. an int16 -2 is being converted to an uint32
			// 4294967294, which is two integer less than 0.
			result = int64(metric.GetIntValue()) - int64(math.MaxUint32) - 1
			return common.ValueTypeInt16, int16(result), nil
		}
		return common.ValueTypeInt16, int16(metric.GetIntValue()), nil
	case protobuf.DataType_Int32:
		var result int64
		if metric.GetIntValue() > math.MaxInt32 {
			// true indicates that the original value is negative, e.g. an int32 -2 is being converted to an uint32
			// 4294967294, which is two integer less than 0.
			result = int64(metric.GetIntValue()) - int64(math.MaxUint32) - 1
			return common.ValueTypeInt32, int32(result), nil
		}
		return common.ValueTypeInt32, int32(metric.GetIntValue()), nil
	case protobuf.DataType_Int64:
		// note that a negative int64 being converted to uint64 can still be directly converted back its original
		// negative int64 value with direct typecast, e.g. an int64 -2 is being converted to an uint64
		// 18446744073709551614, which can be converted back to -2 using typecast int64(uint64(18446744073709551614)).
		// see https://go.dev/play/p/VOM19zwiNtr for a sample code
		return common.ValueTypeInt64, int64(metric.GetIntValue()), nil
	case protobuf.DataType_UInt8:
		return common.ValueTypeUint8, uint8(metric.GetIntValue()), nil
	case protobuf.DataType_UInt16:
		return common.ValueTypeUint16, uint16(metric.GetIntValue()), nil
	case protobuf.DataType_UInt32:
		return common.ValueTypeUint32, uint32(metric.GetLongValue()), nil
	case protobuf.DataType_UInt64:
		return common.ValueTypeUint64, metric.GetLongValue(), nil
	case protobuf.DataType_Float:
		return common.ValueTypeFloat32, metric.GetFloatValue(), nil
	case protobuf.DataType_Double:
		return common.ValueTypeFloat64, metric.GetDoubleValue(), nil
	case protobuf.DataType_Boolean:
		return common.ValueTypeBool, metric.GetBooleanValue(), nil
	case protobuf.DataType_Bytes:
		return common.ValueTypeBinary, metric.GetBytesValue(), nil
	case protobuf.DataType_String, protobuf.DataType_Text:
		return common.ValueTypeString, metric.GetStringValue(), nil
	default:
		return "", nil, fmt.Errorf("metric with unsupported data type %v detected", metric.Datatype)
	}
}
