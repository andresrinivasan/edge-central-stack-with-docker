// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"context"
	"fmt"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	dtosCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCoreCommand(t *testing.T) {
	pushEvent := common.ValueNo
	returnEvent := common.ValueYes
	deviceName := "deviceName"
	goodCommand := "good"
	badCommand := "bad"
	bodyValue := "{\\\"ResourceName\\\":\\\"value\\\"}"
	invalidBodyValue := "I'm not a JSON map"
	testData := "{\"DeviceName\":\"%s\",\"CommandName\":\"%s\",\"RequestMethod\":\"%s\",\"Body\":\"%s\"}"
	settings := map[string]string{"ResourceName": "value"}

	ccMock := &mocks.CommandClient{}
	ccMock.On("IssueGetCommandByName", context.Background(), deviceName, goodCommand, pushEvent, returnEvent).
		Return(&responses.EventResponse{}, nil)
	ccMock.On("IssueGetCommandByName", context.Background(), deviceName, badCommand, pushEvent, returnEvent).
		Return(&responses.EventResponse{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "", nil))
	ccMock.On("IssueSetCommandByName", context.Background(), deviceName, goodCommand, settings).
		Return(dtosCommon.BaseResponse{}, nil)
	ccMock.On("IssueSetCommandByName", context.Background(), deviceName, badCommand, settings).
		Return(dtosCommon.BaseResponse{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "", nil))
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.CommandClientName: func(get di.Get) interface{} {
			return ccMock
		},
	})

	tests := []struct {
		Name                 string
		Data                 interface{}
		Method               string
		DefaultRequestMethod string
		Command              string
		BodyValue            string
		ErrorExpected        bool
		ContinueOnError      bool
	}{
		{"Bad - no data received", nil, methodGet, "", goodCommand, "", true, false},
		{"Good - Get command", testData, methodGet, "", goodCommand, "", false, false},
		{"Bad - Get command (CommandClient returns error)", testData, methodGet, "", badCommand, "", true, false},
		{"Good - Set command", testData, methodSet, "", goodCommand, bodyValue, false, false},
		{"Bad - Set command (CommandClient returns error)", testData, methodSet, "", badCommand, bodyValue, true, false},
		{"Bad - Set command (body is not a JSON map)", testData, methodSet, "", badCommand, invalidBodyValue, true, false},
		{"Bad - request method not found", testData, "", "", goodCommand, "", true, false},
		{"Bad - invalid request method", testData, "Delete", "", goodCommand, "", true, false},
		{"Good - default request method Get", testData, "", methodGet, goodCommand, "", false, false},
		{"Good - default request method Set", testData, "", methodSet, goodCommand, bodyValue, false, false},
		{"Good - ContinueOnError: no data received", nil, methodGet, "", goodCommand, "", true, true},
		{"Good - ContinueOnError: Get command (CommandClient returns error)", testData, methodGet, "", badCommand, "", true, true},
		{"Good - ContinueOnError: Set command (CommandClient returns error)", testData, methodSet, "", badCommand, bodyValue, true, true},
		{"Good - ContinueOnError: Set command (body is not a JSON map)", testData, methodSet, "", badCommand, invalidBodyValue, true, true},
		{"Good - ContinueOnError: request method not found", testData, "", "", goodCommand, "", true, true},
		{"Good - ContinueOnError: invalid request method", testData, "Delete", "", goodCommand, "", true, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			continuePipeline := false
			var result interface{}
			executor := CoreCommandExecutor{
				DeviceNameJSONPath:    "DeviceName",
				CommandNameJSONPath:   "CommandName",
				BodyJSONPath:          "Body",
				RequestMethodJSONPath: "RequestMethod",
				DefaultRequestMethod:  testCase.DefaultRequestMethod,
				PushEvent:             pushEvent,
				ReturnEvent:           returnEvent,
				ContinueOnError:       testCase.ContinueOnError}
			if testCase.Data != nil {
				d := fmt.Sprintf(testCase.Data.(string),
					deviceName, testCase.Command, testCase.Method, testCase.BodyValue)
				continuePipeline, result = executor.Execute(ctx, d)
			} else {
				continuePipeline, result = executor.Execute(ctx, nil)
			}

			if testCase.ErrorExpected {
				if testCase.ContinueOnError {
					require.True(t, continuePipeline)
				} else {
					require.False(t, continuePipeline)
				}
				err := result.(error)
				require.Error(t, err)
				return // Test completed
			}

			require.True(t, continuePipeline)

			switch testCase.Method {
			case methodGet:
				_, ok := result.(*responses.EventResponse)
				assert.True(t, ok)
			case methodSet:
				_, ok := result.(dtosCommon.BaseResponse)
				assert.True(t, ok)
			}
		})
	}
}
