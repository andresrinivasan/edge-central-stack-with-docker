// Copyright (C) 2020-2021 IOTech Ltd

package xpert

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os/exec"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

// Conversion houses various built in conversion transforms (AWSDeviceShadow, JSON, CSV)
type Conversion struct {
}

// NewConversion creates, initializes and returns a new instance of Conversion
func NewConversion() Conversion {
	return Conversion{}
}

// converting event to AWS shadow message in bytes
func convert(param interface{}) ([]byte, error) {
	_, err := json.Marshal(param)
	if err != nil {
		return nil, errors.New(
			"marshaling input data to JSON failed, passed in data must support marshaling to JSON",
		)
	}

	currState := map[string]interface{}{
		"state": map[string]interface{}{
			"reported": param,
		},
	}

	msg, err := json.Marshal(currState)
	if err != nil {
		return []byte{}, err
	}

	return msg, nil
}

func (c Conversion) ConvertToAWSDeviceShadow(ctx interfaces.AppFunctionContext, data interface{}) (
	continuePipeline bool, result interface{}) {
	if data == nil {
		return false, errors.New("no Event Received")
	}

	ctx.LoggingClient().Debug("Transforming to AWS Device Shadow format")

	if event, ok := data.(dtos.Event); ok {
		result, err := convert(event)
		if err != nil {
			ctx.LoggingClient().Errorf("failed to transform to AWS Device Shadow. Error: %s", err)
			return false, err
		}

		return true, result
	}

	return false, errors.New("unexpected type received")
}

// ConvertBoolToIntReading converts readings whose value type is Bool to Int8.  For a Bool reading value whose value is
// true, this function converts its value to 1 in Int8 type.  For a Bool reading value whose value is false, this
// function converts its value to 0 in Int8 type.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
func (c Conversion) ConvertBoolToIntReading(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	ctx.LoggingClient().Debug("Prepare to convert Bool reading to Int reading")
	// Throw out error when no event passed in
	if data == nil {
		return false, errors.New("no Event Received")
	}
	return c.convertBoolToNumericReading(common.ValueTypeInt8, data)
}

// ConvertBoolToFloatReading converts readings whose value type is Bool to Float32.  For a Bool reading value whose value is
// true, this function converts its value to 1 in Float32 type.  For a Bool reading value whose value is false, this
// function converts its value to 0 in Float32 type.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
func (c Conversion) ConvertBoolToFloatReading(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	ctx.LoggingClient().Debug("Prepare to convert Bool reading to Float reading")
	// Throw out error when no event passed in
	if data == nil {
		return false, errors.New("no Event Received")
	}
	return c.convertBoolToNumericReading(common.ValueTypeFloat32, data)
}

func (c Conversion) convertBoolToNumericReading(valueType string, data interface{}) (continuePipeline bool, result interface{}) {
	// ConvertBoolToIntReading expects to deal with event model; for non-event model, simply throw out error.
	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}
	// iterate each reading to convert the Bool reading to target value type
	for i, reading := range event.Readings {
		if reading.ValueType == common.ValueTypeBool {
			event.Readings[i].ValueType = valueType
			boolVal, err := strconv.ParseBool(reading.Value)
			if err != nil {
				return false, fmt.Errorf("failed to parse reading value '%s' to boolean. "+
					"Error:%s", reading.Value, err)
			}
			if boolVal {
				event.Readings[i].Value = BoolTrueNumericValue
			} else {
				event.Readings[i].Value = BoolFalseNumericValue
			}
		}
	}
	return true, event
}

// ConvertIntToFloatReading converts Int or Uint readings to Float readings.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
func (c Conversion) ConvertIntToFloatReading(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	ctx.LoggingClient().Debug("Prepare to convert Int/Uint reading to Float reading")
	// Throw out error when no event passed in
	if data == nil {
		return false, errors.New("no Event Received")
	}
	// ConvertIntToFloatReading expects to deal with event dto; for non-event dto, simply throw out error.
	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// iterate each reading to convert the Int reading or Uint reading to Float reading
	for i, reading := range event.Readings {
		switch reading.ValueType {
		case common.ValueTypeInt8, common.ValueTypeInt16, common.ValueTypeInt32, common.ValueTypeInt64, common.ValueTypeUint8, common.ValueTypeUint16, common.ValueTypeUint32, common.ValueTypeUint64:
			ctx.LoggingClient().Debugf("Convert reading with value type %s to Float reading", reading.ValueType)
			f, err := strconv.ParseFloat(reading.Value, 64)
			if err != nil {
				return false, fmt.Errorf("failed to parse reading value '%s' to float. "+
					"Error:%s", reading.Value, err)
			} else {
				event.Readings[i].ValueType = common.ValueTypeFloat64
				event.Readings[i].Value = fmt.Sprintf("%e", f)
			}
		default:
			ctx.LoggingClient().Debugf("Skip conversion for value type %s reading", reading.ValueType)
			// ignore other types
			continue
		}
	}
	return true, event
}

// ConvertFloatToIntReading converts readings whose value type is Float to Int.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
func (c Conversion) ConvertFloatToIntReading(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	ctx.LoggingClient().Debug("Prepare to convert Float reading to Int reading")
	// Throw out error when no event passed in
	if data == nil {
		return false, errors.New("no Event Received")
	}
	// ConvertFloatToIntReading expects to deal with event dto; for non-event dto, simply throw out error.
	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// iterate each reading to convert the Float reading to target value type
	for i, reading := range event.Readings {
		var f float64
		switch reading.ValueType {
		case common.ValueTypeFloat32, common.ValueTypeFloat64:
			var err error
			// strconv.ParseFloat returns the "nearest" floating-point number rounded using IEEE754 unbiased rounding.
			// The possible consequence of using strconv.ParseFloat is the loss of precision.
			f, err = strconv.ParseFloat(reading.Value, 64)
			if err != nil {
				return false, fmt.Errorf("failed to parse reading value '%s' to Int. "+
					"Error:%s", reading.Value, err)
			}
		default:
			// ignore other types
			continue
		}
		if big.NewFloat(f).Cmp(big.NewFloat(float64(math.MaxInt64))) > 0 ||
			big.NewFloat(f).Cmp(big.NewFloat(float64(math.MinInt64))) < 0 {
			return false, fmt.Errorf("value %e is outside the range of Int64", f)
		}

		event.Readings[i].ValueType = common.ValueTypeInt64
		event.Readings[i].Value = fmt.Sprintf("%d", int64(f))
	}
	return true, event
}

func (c Conversion) ConvertByteArrayToEvent(ctx interfaces.AppFunctionContext, data interface{}) (
	continuePipeline bool, result interface{}) {
	if data == nil {
		return false, errors.New("no data Received")
	}

	ctx.LoggingClient().Debug("Transforming to event DTO")

	var err error
	switch d := data.(type) {
	case []byte:
		var event dtos.Event
		if err = json.Unmarshal(d, &event); err == nil {
			result = event
		}
	default: // receive unsupported data type
		err = fmt.Errorf("only support the conversion from byte array")
	}

	if err != nil {
		ctx.LoggingClient().Errorf("type received is not an Event. Error: %s", err)
		return false, err
	} else {
		return true, result
	}
}

const (
	NodeJSCommand             = "node"
	placeHolderInputString    = "$$INPUTJSONSTRING$$"
	placeHolderTransformLogic = "$$TRANSFORMLOGIC$$"
	jsTransformTemplate       = "var inputJsonString = `" + placeHolderInputString + "`\nvar input = JSON.parse(inputJsonString);\nfunction xpertAppTransform (inputObject){\n    " + placeHolderTransformLogic + "\n}\nprocess.stdout.write(JSON.stringify(xpertAppTransform(input)))"
)

// ScriptTransform will transform the incoming data to specified format per provided JavaScript
type ScriptTransform struct {
	script string
}

// NewScriptTransform creates, initializes and returns a new instance of NewScriptTransform
func NewScriptTransform(transformScript string) ScriptTransform {
	return ScriptTransform{
		script: transformScript,
	}
}

// Transform run specified javascript against incoming data.
// The incoming data will be accessible as inputObject, and the conversion result must be returned if the conversion
// result needs to be passed next function in the pipeline.
// This function will return an error and stop the pipeline if any error occurs when running specified script or if no
// data is received. This function is a configuration function and returns a function pointer.
func (st ScriptTransform) Transform(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	ctx.LoggingClient().Debugf("Prepare to run JavascriptTransform in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		// didn't receive any result
		return false, fmt.Errorf("function JavascriptTransform in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	byteData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}
	inputJsonString := string(byteData)
	ctx.LoggingClient().Debugf("JavascriptTransform in pipeline '%s' receives data: %s", ctx.PipelineId(), inputJsonString)

	// following replacements will inject inputJsonString and transform logic into a javascript snippet as shown below:
	// var inputJsonString = '$$INPUTJSONSTRING$$'
	// var input = JSON.parse(inputJsonString);
	// function xpertAppTransform (inputObject){
	//    $$TRANSFORMLOGIC$$
	// }
	// process.stdout.write(JSON.stringify(xpertAppTransform(input)))
	jsTransformScript := strings.Replace(jsTransformTemplate, placeHolderInputString, inputJsonString, 1)
	jsTransformScript = strings.Replace(jsTransformScript, placeHolderTransformLogic, st.script, 1)

	ctx.LoggingClient().Debugf("prepare to run transform script: %s", jsTransformScript)
	output, err := runTransformScript(jsTransformScript)
	if err != nil {
		return false, fmt.Errorf("error occurs while executing transfrom script:%s. Error: %s", jsTransformScript, err.Error())
	}

	ctx.LoggingClient().Debugf("result of running script: %s", string(output))
	return true, output
}

func runTransformScript(scriptContent string) ([]byte, error) {
	// delegate the scriptContent as stdin of nodejs execution, note that this piece of code assume that app-service has nodejs installed
	command := exec.Command(NodeJSCommand)
	command.Stdin = strings.NewReader(scriptContent)
	output, err := command.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			// if it's ExitError, reformat the error with proper messages.
			return nil, errors.New(string(ee.Stderr))
		}
		return nil, err
	}
	return output, nil
}
