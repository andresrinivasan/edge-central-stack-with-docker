// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"context"
	"encoding/json"
	"fmt"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"net/http"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/tidwall/gjson"
)

const (
	methodGet = http.MethodGet
	methodSet = "SET"
)

type CoreCommandExecutor struct {
	DeviceNameJSONPath    string
	CommandNameJSONPath   string
	BodyJSONPath          string
	RequestMethodJSONPath string
	DefaultRequestMethod  string
	PushEvent             string
	ReturnEvent           string
	ContinueOnError       bool
}

// Execute parses the incoming JSON data to determine the core command and then issues it through the CommandClient.
// The result to be returned depends on the given request method: EventResponse for Get and BaseResponse for Set.
// When an error occurs, this function returns the error in the type of EdgeX.
func (executor CoreCommandExecutor) Execute(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return executor.ContinueOnError, edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid,
			"no Data Received", nil)
	}

	byteData, err := util.CoerceType(data)
	if err != nil {
		return executor.ContinueOnError, edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid,
			"invalid data type", err)
	}

	method := gjson.Get(string(byteData), executor.RequestMethodJSONPath).Str
	if len(method) == 0 {
		if len(executor.DefaultRequestMethod) == 0 {
			return executor.ContinueOnError, edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid,
				"DefaultRequestMethod is required when RequestMethodJSONPath is not set", nil)
		}
		method = executor.DefaultRequestMethod
	}
	deviceName := gjson.Get(string(byteData), executor.DeviceNameJSONPath).Str
	commandName := gjson.Get(string(byteData), executor.CommandNameJSONPath).Str

	switch strings.ToUpper(method) {
	case methodSet:
		body := gjson.Get(string(byteData), executor.BodyJSONPath).Str
		var settings map[string]string
		err = json.Unmarshal([]byte(body), &settings)
		if err != nil {
			return executor.ContinueOnError, edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, "could not parse the input body to map[string]string", err)
		}
		baseResponse, err := ctx.CommandClient().IssueSetCommandByName(context.Background(), deviceName, commandName, settings)
		if err != nil {
			return executor.ContinueOnError, err
		} else {
			return true, baseResponse
		}
	case methodGet:
		eventResponse, err := ctx.CommandClient().IssueGetCommandByName(context.Background(), deviceName, commandName, executor.PushEvent, executor.ReturnEvent)
		if err != nil {
			return executor.ContinueOnError, err
		} else {
			return true, eventResponse
		}
	default:
		return executor.ContinueOnError, edgexErrors.NewCommonEdgeX(edgexErrors.KindNotAllowed,
			fmt.Sprintf("the request method %s is not supported", method), nil)
	}
}
