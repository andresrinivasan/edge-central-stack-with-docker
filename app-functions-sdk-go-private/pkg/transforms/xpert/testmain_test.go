// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"os"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	contractCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	contractDtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/stretchr/testify/mock"
)

var lc logger.LoggingClient
var dic *di.Container
var ctx *appfunction.Context

func TestMain(m *testing.M) {
	lc = logger.NewMockClient()
	config := &common.ConfigurationStruct{}

	dic = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return config
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	initProfileClient()

	ctx = appfunction.NewContext("123", dic, "")

	os.Exit(m.Run())
}

func initProfileClient() {
	resources := []dtos.DeviceResource{
		{Name: contractCommon.ValueTypeInt8, Properties: dtos.ResourceProperties{Minimum: IntMIN, Maximum: IntMAX, ValueType: contractCommon.ValueTypeInt8}},
		{Name: contractCommon.ValueTypeInt16, Properties: dtos.ResourceProperties{Minimum: IntMIN, Maximum: IntMAX, ValueType: contractCommon.ValueTypeInt16}},
		{Name: contractCommon.ValueTypeInt32, Properties: dtos.ResourceProperties{Minimum: IntMIN, Maximum: IntMAX, ValueType: contractCommon.ValueTypeInt32}},
		{Name: contractCommon.ValueTypeInt64, Properties: dtos.ResourceProperties{Minimum: IntMIN, Maximum: IntMAX, ValueType: contractCommon.ValueTypeInt64}},
		{Name: contractCommon.ValueTypeUint8, Properties: dtos.ResourceProperties{Minimum: UintMIN, Maximum: UintMAX, ValueType: contractCommon.ValueTypeUint8}},
		{Name: contractCommon.ValueTypeUint16, Properties: dtos.ResourceProperties{Minimum: UintMIN, Maximum: UintMAX, ValueType: contractCommon.ValueTypeUint16}},
		{Name: contractCommon.ValueTypeUint32, Properties: dtos.ResourceProperties{Minimum: UintMIN, Maximum: UintMAX, ValueType: contractCommon.ValueTypeUint32}},
		{Name: contractCommon.ValueTypeUint64, Properties: dtos.ResourceProperties{Minimum: UintMIN, Maximum: UintMAX, ValueType: contractCommon.ValueTypeUint64}},
		{Name: contractCommon.ValueTypeFloat32, Properties: dtos.ResourceProperties{Minimum: Float32MIN, Maximum: Float32MAX, ValueType: contractCommon.ValueTypeFloat32}},
		{Name: contractCommon.ValueTypeFloat64, Properties: dtos.ResourceProperties{Minimum: Float64MIN, Maximum: Float64MAX, ValueType: contractCommon.ValueTypeFloat64}},
		{Name: "test", Tags: tagsToAdd},
	}
	mockdpc := mocks.DeviceProfileClient{}
	mockdpc.On("AllDeviceProfiles", mock.Anything, mock.Anything, 0, -1).Return(responses.MultiDeviceProfilesResponse{
		BaseWithTotalCountResponse: contractDtoCommon.BaseWithTotalCountResponse{},
		Profiles: []dtos.DeviceProfile{{
			DeviceResources: resources,
		}},
	}, nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return &mockdpc
		},
	})
}
