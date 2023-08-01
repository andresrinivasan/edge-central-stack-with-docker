// Copyright (C) 2022 IOTech Ltd

package xpert

import (
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
	"github.com/tidwall/gjson"
)

// ContextVariableSetter is the setter used to set context variable
type ContextVariableSetter struct {
	VariableName    string
	ValueJsonPath   string
	ContinueOnError bool
}

func NewContextVariableSetter(variableName, valueJsonPath string, continueOnError bool) ContextVariableSetter {
	return ContextVariableSetter{variableName, valueJsonPath, continueOnError}
}

// SetContextVariable sets the context variable per specified parameters.  The name of context variable will be the
// value of parameter VariableName, and the value of context variable will be the value extracted from the specified
// JsonPath of data as passed into the function.  Set continueOnError to true if users would like the pipeline to
// continue even when there is error during SetContextVariable.  When continueOnError is true and error occurs during
// SetContextVariable, the function will return incoming data rather than error. Please note that if incoming data
// doesn't contain specified JsonPath, the context variable will be set to empty string "".
// This function is a configuration function and returns a function pointer.
func (s ContextVariableSetter) SetContextVariable(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {

	ctx.LoggingClient().Debugf("Setting context variable %s in pipeline '%s'", s.VariableName, ctx.PipelineId())

	var err error
	var byteData []byte

	if data == nil {
		// didn't receive a result
		err = fmt.Errorf("function SetContextVariable in pipeline '%s': No Data Received", ctx.PipelineId())
	} else {
		// attempt to convert non-nil incoming data into bytes
		byteData, err = util.CoerceType(data)
		if err == nil {
			// check if incoming data is in well-formed json
			if !gjson.ValidBytes(byteData) {
				err = fmt.Errorf("function SetContextVariable in pipeline '%s': incoming data is not in valid json format: %s", ctx.PipelineId(), string(byteData))
			}
		}
	}

	if err != nil {
		ctx.LoggingClient().Error(err.Error())
		if s.ContinueOnError {
			// under error case, return the incoming data when continueOnError is true
			return true, data
		} else {
			// under error case, return error when continueOnError is false
			return false, err
		}
	}

	// note that if byteData doesn't have corresponding jsonPath, the value will be empty string ""
	value := gjson.Get(string(byteData), s.ValueJsonPath).String()
	ctx.AddValue(s.VariableName, value)

	return true, data
}
