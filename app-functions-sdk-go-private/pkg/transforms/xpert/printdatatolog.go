// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
)

// PrintDataToLog prints the incoming data for debugging purpose.
func PrintDataToLog(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	input, err := util.CoerceType(data)
	if err != nil {
		ctx.LoggingClient().Error(err.Error())
		return false, err
	}
	ctx.LoggingClient().Infof("Data: %s", string(input))
	return true, data
}
