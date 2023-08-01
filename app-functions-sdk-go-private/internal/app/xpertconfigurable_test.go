// Copyright (C) 2021-2023 IOTech Ltd

package app

import (
	"net/http"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/stretchr/testify/assert"
)

func TestXpertMQTTExportNoExportMode(t *testing.T) {
	configurable := Configurable{lc: lc}
	params := make(map[string]string)
	// no ExportMode specified in params
	trx := configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, "return result from XpertMQTTExport should be nil")
}

func TestXpertMQTTExportCommon(t *testing.T) {
	configurable := Configurable{lc: lc}
	params := make(map[string]string)
	msgResultShouldBeNil := "return result from XpertMQTTExport should be nil"
	msgResultShouldNotBeNil := "return result from XpertMQTTExport shouldn't be nil"

	// only export mode
	params[ExportMode] = RegularMQTT
	trx := configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// export mode and broker address
	params[BrokerAddress] = "tls://test.mosquitto.org:8883"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// export mode, broker address, and MQTT topic
	params[Topic] = "edgexpert/events"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// export mode, broker address, MQTT topic and auth mode; all satisfied
	params[AuthMode] = messaging.AuthModeNone
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable Qos
	params[Qos] = "test"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable, invalid Qos
	params[Qos] = "3"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable, valid Qos
	params[Qos] = "0"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// parsable Retain
	params[Retain] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable AutoReconnect
	params[AutoReconnect] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable AutoReconnect
	params[AutoReconnect] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable SkipVerify
	params[SkipVerify] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable SkipVerify
	params[SkipVerify] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable PersistOnError
	params[PersistOnError] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable PersistOnError
	params[PersistOnError] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// selected auth mod requires secrets, but secret path is not specified
	params[AuthMode] = messaging.AuthModeCert
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// secret path is specified
	params[SecretPath] = "mypath"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)
}

func TestXpertMQTTExportAWS(t *testing.T) {
	configurable := Configurable{lc: lc}
	params := make(map[string]string)
	msgResultShouldBeNil := "return result from XpertMQTTExport should be nil"
	msgResultShouldNotBeNil := "return result from XpertMQTTExport shouldn't be nil"

	// only export mode
	params[ExportMode] = AWSIoTCore
	trx := configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// only broker address
	params[BrokerAddress] = "tls://a1hmdzcnpgcx08-ats.iot.us-east-1.amazonaws.com:8883"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// only broker address and MQTT topic
	params[Topic] = "$aws/things/VirtualDevice/shadow/update"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// broker address, port, and client ID
	params[ClientID] = "VirtualDevice"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// broker address, port, client ID, and secretPath; all satisfied
	params[SecretPath] = "aws"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable Qos
	params[Qos] = "test"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable, invalid Qos
	params[Qos] = "2"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable, valid Qos
	params[Qos] = "0"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// parsable Retain
	params[Retain] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable AutoReconnect
	params[AutoReconnect] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable AutoReconnect
	params[AutoReconnect] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable SkipVerify
	params[SkipVerify] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable SkipVerify
	params[SkipVerify] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable PersistOnError
	params[PersistOnError] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable PersistOnError
	params[PersistOnError] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)
}

func TestXpertMQTTExportAzure(t *testing.T) {
	configurable := Configurable{lc: lc}
	params := make(map[string]string)
	msgResultShouldBeNil := "return result from XpertMQTTExport should be nil"
	msgResultShouldNotBeNil := "return result from XpertMQTTExport shouldn't be nil"

	// only export mode
	params[ExportMode] = AzureIoTHub
	trx := configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// only broker address
	params[BrokerAddress] = "tls://EdgeX.azure-devices.net:8883"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// only broker address and topic
	params[Topic] = "devices/TestDevice01/messages/events/"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// broker address, topic, and client ID
	params[ClientID] = "test"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// broker address, topic, client ID, and secret path
	params[SecretPath] = "iothub"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// broker address, topic, client ID, secret path and auth mode; all mandatory configuration satisfied
	params[AuthMode] = "clientcert"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable Qos
	params[Qos] = "test"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable Qos
	params[Qos] = "0"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable Retain
	params[Retain] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable Retain
	params[Retain] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable AutoReconnect
	params[AutoReconnect] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable AutoReconnect
	params[AutoReconnect] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable SkipVerify
	params[SkipVerify] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable SkipVerify
	params[SkipVerify] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unsupported authMode
	params[AuthMode] = "cacert"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)
}

func TestXpertMQTTExportIBM(t *testing.T) {
	configurable := Configurable{lc: lc}
	params := make(map[string]string)
	msgResultShouldBeNil := "return result from XpertMQTTExport should be nil"
	msgResultShouldNotBeNil := "return result from XpertMQTTExport shouldn't be nil"

	// no mandatory configuration specified in params
	trx := configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// only export mode
	params[ExportMode] = IBMWatson
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// only broker address
	params[BrokerAddress] = "tcps://3wrbbc.messaging.internetofthings.ibmcloud.com:8883"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// only broker address and topic
	params[Topic] = "iot-2/evt/mesaage123/fmt/json"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// broker address, topic, and clientid
	params[ClientID] = "d:3wrbbc:edgex:test01"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// broker address, topic, clientid, and secretPath; all satisfied
	params[SecretPath] = "watson"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsible Qos
	params[Qos] = "test"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsible Qos
	params[Qos] = "0"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsible Retain
	params[Retain] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsible Retain
	params[Retain] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsible AutoReconnect
	params[AutoReconnect] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsible AutoReconnect
	params[AutoReconnect] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsible SkipVerify
	params[SkipVerify] = "ttt"
	trx = configurable.XpertMQTTExport(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsible SkipVerify
	params[SkipVerify] = "true"
	trx = configurable.XpertMQTTExport(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)
}

func TestConvertToAWSDeviceShadow(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.ConvertToAWSDeviceShadow()
	assert.NotNil(t, trx, "return result from TransformToAWSDeviceShadow should not be nil")
}

func TestConfigurableConvertBoolToIntReading(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.ConvertBoolToIntReading()
	assert.NotNil(t, trx, "return result from ConvertBoolToIntReading should not be nil")
}

func TestConfigurableConvertBoolToFloatReading(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.ConvertBoolToFloatReading()
	assert.NotNil(t, trx, "return result from ConvertBoolToFloatReading should not be nil")
}

func TestConfigurableConvertIntToFloatReading(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.ConvertIntToFloatReading()
	assert.NotNil(t, trx, "return result from ConvertIntToFloatReading should not be nil")
}

func TestConfigurableConvertFloatToIntReading(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.ConvertFloatToIntReading()
	assert.NotNil(t, trx, "return result from ConvertFloatToIntReading should not be nil")
}

func ConvertByteArrayToEvent(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.ConvertByteArrayToEvent()
	assert.NotNil(t, trx, "return result from ConvertByteArrayToEvent should not be nil")
}

func TestConfigurableFilterByValueMaxMin(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.FilterByValueMaxMin()
	assert.NotNil(t, trx, "return result from FilterByValueMaxMin should not be nil")
}

func TestConfigurableXpertAddTagsFromDeviceResource(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.XpertAddTagsFromDeviceResource()
	assert.NotNil(t, trx, "return result from AddTagsFromDeviceResource should not be nil")
}

func TestConfigurableXpertAddTagsFromDevice(t *testing.T) {
	configurable := Configurable{lc: lc}
	trx := configurable.XpertAddTagsFromDevice()
	assert.NotNil(t, trx, "return result from AddTagsFromDevice should not be nil")
}

func TestSetContextVariable(t *testing.T) {
	configurable := Configurable{lc: lc}

	testVariableName := "testVar"
	testValueJsonPath := "a.b.c"
	validBoolValue := "false"
	invalidBoolValue := "bogus"

	tests := []struct {
		Name            string
		VariableName    *string
		ValueJsonPath   *string
		ContinueOnError *string
		ExpectValid     bool
	}{
		{"Valid - with required params", &testVariableName, &testValueJsonPath, nil, true},
		{"Valid - with all params", &testVariableName, &testValueJsonPath, &validBoolValue, true},
		{"Invalid - no variableName", nil, &testValueJsonPath, nil, false},
		{"Invalid - no valueJsonPath", &testVariableName, nil, nil, false},
		{"Invalid - bad ContinueOnSendError", &testVariableName, &testValueJsonPath, &invalidBoolValue, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			params := make(map[string]string)

			if test.VariableName != nil {
				params[VariableName] = *test.VariableName
			}

			if test.ValueJsonPath != nil {
				params[ValueJSONPath] = *test.ValueJsonPath
			}

			if test.ContinueOnError != nil {
				params[continueOnError] = *test.ContinueOnError
			}

			setter := configurable.SetContextVariable(params)
			assert.Equal(t, test.ExpectValid, setter != nil)
		})
	}
}

func TestXpertHTTPExport(t *testing.T) {
	configurable := Configurable{lc: lc}

	testUrl := "http://url"
	testMimeType := common.ContentTypeJSON
	validBoolValue := "false"
	invalidBoolValue := "bogus"
	invalidRenegotiationSupportValue := "3"

	testHeaderName := "My-Header"
	testSecretPath := "/path"
	testSecretName := "header"

	testAuthModeHeaderSecret := string(xpert.HTTPAuthModeHeaderSecret)
	testAuthModeOAuth2CC := string(xpert.HTTPAuthModeOauth2ClientCredentials)
	testAuthModeClientCert := string(xpert.HTTPAuthModeClientCert)

	testHTTPRequestHeaders := "{\"Accept\": \"text/plain\"}"
	invalidHTTPRequestHeaders := "0"
	testEmptyHTTPRequestHeaders := ""

	tests := []struct {
		Name                 string
		Method               string
		Url                  *string
		MimeType             *string
		PersistOnError       *string
		ContinueOnSendError  *string
		ReturnInputData      *string
		HeaderName           *string
		SecretPath           *string
		SecretName           *string
		AuthMode             *string
		RenegotiationSupport *string
		HTTPRequestHeaders   *string
		ExpectValid          bool
	}{
		{"Valid - only required params", http.MethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, nil, nil, nil, true},
		{"Valid - empty HTTPRequestHeaders", http.MethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, nil, nil, &testEmptyHTTPRequestHeaders, true},
		{"Valid - w/o secrets", http.MethodPut, &testUrl, &testMimeType, &validBoolValue, nil, nil, nil, nil, nil, nil, nil, nil, true},
		{"Valid - with secrets", http.MethodPatch, &testUrl, &testMimeType, nil, nil, nil, &testHeaderName, &testSecretPath, &testSecretName, &testAuthModeHeaderSecret, nil, nil, true},
		{"Valid - with all params", http.MethodDelete, &testUrl, &testMimeType, &validBoolValue, &validBoolValue, &validBoolValue, &testHeaderName, &testSecretPath, &testSecretName, &testAuthModeHeaderSecret, nil, &testHTTPRequestHeaders, true},
		{"Invalid - unsupported http method", http.MethodOptions, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid - no url", http.MethodPost, nil, &testMimeType, nil, nil, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid - no mimeType", http.MethodPost, &testUrl, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid - bad persistOnError", http.MethodPost, &testUrl, &testMimeType, &invalidBoolValue, nil, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid - bad continueOnSendError", http.MethodPost, &testUrl, &testMimeType, nil, &invalidBoolValue, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid - bad returnInputData", http.MethodPost, &testUrl, &testMimeType, nil, nil, &invalidBoolValue, nil, nil, nil, nil, nil, nil, false},
		{"Invalid - header secret & missing headerName", http.MethodPost, &testUrl, &testMimeType, &validBoolValue, nil, nil, nil, &testSecretPath, &testSecretName, &testAuthModeHeaderSecret, nil, nil, false},
		{"Invalid - header secret & missing secretPath", http.MethodPost, &testUrl, &testMimeType, &validBoolValue, nil, nil, &testHeaderName, nil, &testSecretName, &testAuthModeHeaderSecret, nil, nil, false},
		{"Invalid - header secret & missing secretName", http.MethodPost, &testUrl, &testMimeType, &validBoolValue, nil, nil, &testHeaderName, &testSecretPath, nil, &testAuthModeHeaderSecret, nil, nil, false},
		{"Invalid - oauth2 & missing secretPath", http.MethodPost, &testUrl, &testMimeType, &validBoolValue, nil, nil, nil, nil, nil, &testAuthModeOAuth2CC, nil, nil, false},
		{"Invalid - clientcert & missing secretPath", http.MethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, &testAuthModeClientCert, nil, nil, false},
		{"Invalid - unsupported renegotiation", http.MethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, &testSecretPath, nil, &testAuthModeClientCert, &invalidRenegotiationSupportValue, nil, false},
		{"Invalid - invalid http request headers", http.MethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, &testSecretPath, nil, &testAuthModeClientCert, nil, &invalidHTTPRequestHeaders, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			params := make(map[string]string)
			params[ExportMethod] = test.Method

			if test.Url != nil {
				params[Url] = *test.Url
			}

			if test.MimeType != nil {
				params[MimeType] = *test.MimeType
			}

			if test.PersistOnError != nil {
				params[PersistOnError] = *test.PersistOnError
			}

			if test.ContinueOnSendError != nil {
				params[ContinueOnSendError] = *test.ContinueOnSendError
			}

			if test.ReturnInputData != nil {
				params[ReturnInputData] = *test.ReturnInputData
			}

			if test.HeaderName != nil {
				params[HeaderName] = *test.HeaderName
			}

			if test.SecretPath != nil {
				params[SecretPath] = *test.SecretPath
			}

			if test.SecretName != nil {
				params[SecretName] = *test.SecretName
			}

			if test.AuthMode != nil {
				params[AuthMode] = *test.AuthMode
			}

			if test.RenegotiationSupport != nil {
				params[RenegotiationSupport] = *test.RenegotiationSupport
			}

			if test.HTTPRequestHeaders != nil {
				params[HTTPRequestHeaders] = *test.HTTPRequestHeaders
			}

			transform := configurable.XpertHTTPExport(params)
			assert.Equal(t, test.ExpectValid, transform != nil)
		})
	}
}

func TestXpertHTTPExportAWSSignature(t *testing.T) {
	configurable := Configurable{lc: lc}

	testSecretPath := "aws"
	testAWSV4SignerConfigs := "{\"region\": \"us-east-1\", \"service\": \"s3\"}"
	invalidAWSV4SignerConfigsRegion := "{\"service\": \"s3\"}"
	invalidAWSV4SignerConfigsService := "{\"region\": \"us-east-1\"}"
	invalidEmptyAWSV4SignerConfigsRegion := "{\"region\": \"\", \"service\": \"s3\"}"
	invalidEmptyAWSV4SignerConfigsService := "{\"region\": \"us-east-1\", \"service\": \"\"}"

	tests := []struct {
		Name               string
		AWSV4SignerConfigs *string
		SecretPath         *string
		ExpectValid        bool
	}{
		{"Valid", &testAWSV4SignerConfigs, &testSecretPath, true},
		{"Invalid - missing AWSV4SignerConfigs", nil, &testSecretPath, false},
		{"Invalid - missing region", &invalidAWSV4SignerConfigsRegion, &testSecretPath, false},
		{"Invalid - empty region", &invalidEmptyAWSV4SignerConfigsRegion, &testSecretPath, false},
		{"Invalid - missing service", &invalidAWSV4SignerConfigsService, &testSecretPath, false},
		{"Invalid - empty service", &invalidEmptyAWSV4SignerConfigsService, &testSecretPath, false},
		{"Invalid - missing secret path", &testAWSV4SignerConfigs, nil, false},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			params := make(map[string]string)
			params[Url] = "http://url"
			params[MimeType] = common.ContentTypeJSON
			params[ExportMethod] = "GET"
			params[AuthMode] = string(xpert.HTTPAuthModeAWSSignature)

			if test.AWSV4SignerConfigs != nil {
				params[AWSV4SignerConfigs] = *test.AWSV4SignerConfigs
			}
			if test.SecretPath != nil {
				params[SecretPath] = *test.SecretPath
			}
			transform := configurable.XpertHTTPExport(params)
			assert.Equal(t, test.ExpectValid, transform != nil)
		})
	}
}

func TestJavascriptTransform(t *testing.T) {
	configurable := Configurable{lc: lc}

	testScript := "var test"
	emptyScript := ""

	tests := []struct {
		Name        string
		Script      *string
		ExpectValid bool
	}{
		{"Valid - with required params", &testScript, true},
		{"Invalid - empty script", &emptyScript, false},
		{"Invalid - no script", nil, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			params := make(map[string]string)

			if test.Script != nil {
				params[TransformScript] = *test.Script
			}

			transform := configurable.JavascriptTransform(params)
			assert.Equal(t, test.ExpectValid, transform != nil)
		})
	}
}

func TestConvertDDATAToEvent(t *testing.T) {
	configurable := Configurable{lc: lc}
	msgResultShouldNotBeNil := "return result from ConvertDDATAToEvent shouldn't be nil"
	trx := configurable.ConvertDDATAToEvent()
	assert.NotNil(t, trx, msgResultShouldNotBeNil)
}
