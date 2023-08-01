// Copyright (C) 2021-2023 IOTech Ltd

package app

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
)

const (
	ExportMode      = "exportmode"
	UseSharedClient = "usesharedclient"

	RegularMQTT               = "mqtt"
	AWSIoTCore                = "awsiotcore"
	AWSIoTCoreResponse        = "awsiotcoreresponse"
	AzureIoTHub               = "azureiothub"
	AzureDirectMethodResponse = "azuredirectmethodresponse"
	IBMWatson                 = "ibmwatson"

	VariableName  = "variablename"
	ValueJSONPath = "valuejsonpath"

	TransformScript = "transformscript"

	Namespace  = "namespace"
	GroupId    = "groupid"
	EdgeNodeId = "edgenodeid"

	RenegotiationSupport = "renegotiationsupport"
	HTTPRequestHeaders   = "httprequestheaders"
	AWSV4SignerConfigs   = "awsv4signerconfigs"
)

func (app *Configurable) PrintDataToLog() interfaces.AppFunction {
	return xpert.PrintDataToLog
}

// XpertMQTTExport will send data from the previous function to the specified Endpoint via MQTT publish. If no previous function exists,
// then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) XpertMQTTExport(parameters map[string]string) interfaces.AppFunction {
	mode, ok := parameters[ExportMode]
	if !ok {
		app.lc.Errorf("Could not find '%s' parameter for Transform", ExportMode)
		return nil
	}
	var err error
	useSharedClient := false
	if value, ok := parameters[UseSharedClient]; ok {
		useSharedClient, err = strconv.ParseBool(value)
		if err != nil {
			app.lc.Errorf("could not parse '%s' to a bool for '%s' parameter. ", value, UseSharedClient)
			return nil
		}
	}

	switch strings.ToLower(mode) {
	case RegularMQTT:
		if useSharedClient {
			parameters = LoadSharedMQTTConfig(app.externalMqttConfig, parameters)
		}
		config, persistOnError, err := LoadRegularMQTTConfig(app.lc, parameters)
		if err != nil {
			app.lc.Errorf("failed to load configuration, error: %s", err)
			return nil
		}
		sender := xpert.NewRegularMQTTSender(config, persistOnError, useSharedClient)
		return sender.Send
	case AWSIoTCore:
		if useSharedClient {
			parameters = LoadSharedMQTTConfig(app.externalMqttConfig, parameters)
		}
		config, persistOnError, err := LoadAWSMQTTConfig(app.lc, parameters)
		if err != nil {
			app.lc.Errorf("failed to load configuration, error: %s", err)
			return nil
		}
		sender := xpert.NewAWSIoTCoreMQTTSender(config, persistOnError, useSharedClient)
		return sender.Send
	case AWSIoTCoreResponse:
		parameters = LoadSharedMQTTConfig(app.externalMqttConfig, parameters)
		config, persistOnError, err := LoadAWSMQTTConfig(app.lc, parameters)
		if err != nil {
			app.lc.Errorf("failed to load configuration, error: %s", err)
			return nil
		}
		sender := xpert.NewAWSIoTCoreMQTTSender(config, persistOnError, true)
		return sender.SendResponse
	case AzureIoTHub:
		if useSharedClient {
			parameters = LoadSharedMQTTConfig(app.externalMqttConfig, parameters)
		}
		config, persistOnError, err := LoadAzureMQTTConfig(app.lc, parameters)
		if err != nil {
			app.lc.Errorf("failed to load configuration, error: %s", err)
			return nil
		}
		sender := xpert.NewAzureIoTHubMQTTSender(config, persistOnError, useSharedClient)
		return sender.Send
	case AzureDirectMethodResponse:
		// Put a necessary key to the parameters map in order to pass loadAzureMQTTConfig func.
		// The actual topic will be determined right before sending the response.
		parameters[Topic] = ""
		parameters = LoadSharedMQTTConfig(app.externalMqttConfig, parameters)
		config, persistOnError, err := LoadAzureMQTTConfig(app.lc, parameters)
		if err != nil {
			app.lc.Errorf("failed to load configuration, error: %s", err)
			return nil
		}
		sender := xpert.NewAzureIoTHubMQTTSender(config, persistOnError, true)
		return sender.SendDirectMethodResponse
	case IBMWatson: // TODO: Remove this case in v3.0
		if useSharedClient {
			parameters = LoadSharedMQTTConfig(app.externalMqttConfig, parameters)
		}
		config, persistOnError, err := LoadIBMMQTTConfig(app.lc, parameters)
		if err != nil {
			app.lc.Errorf("failed to load configuration, error: %s", err)
			return nil
		}
		sender := xpert.NewIBMWatsonMQTTSender(config, persistOnError, useSharedClient) // nolint: staticcheck
		return sender.Send
	default:
		app.lc.Errorf(
			"invalid export mode '%s'", mode)
		app.lc.Infof("Valid export mode regarding common MQTT broker: '%s'", RegularMQTT)
		app.lc.Infof("Valid export modes regarding AWS IoT Core: '%s' and '%s'",
			AWSIoTCore,
			AWSIoTCoreResponse)
		app.lc.Infof("Valid export modes regarding Azure IoT Hub: '%s' and '%s'",
			AzureIoTHub,
			AzureDirectMethodResponse)
		app.lc.Infof("Valid export mode regarding IBM Watson: '%s'", IBMWatson)
		return nil
	}
}

// ConvertToAWSDeviceShadow converts an EdgeX Event to AWS Device Shadow document.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) ConvertToAWSDeviceShadow() interfaces.AppFunction {
	conversion := xpert.Conversion{}
	return conversion.ConvertToAWSDeviceShadow
}

// ConvertBoolToIntReading converts readings whose value type is Bool to Int8.  For a Bool reading value whose value is
// true, this function converts its value to 1 in Int8 type.  For a Bool reading value whose value is false, this
// function converts its value to 0 in Int8 type.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) ConvertBoolToIntReading() interfaces.AppFunction {
	conversion := xpert.Conversion{}
	return conversion.ConvertBoolToIntReading
}

// ConvertBoolToFloatReading converts readings whose value type is Bool to Float32.  For a Bool reading value whose value is
// true, this function converts its value to 1 in Float32 type.  For a Bool reading value whose value is false, this
// function converts its value to 0 in Float32 type.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) ConvertBoolToFloatReading() interfaces.AppFunction {
	conversion := xpert.Conversion{}
	return conversion.ConvertBoolToFloatReading
}

// ConvertIntToFloatReading converts readings whose value type is among Int8, Int16, Int32, or Int64 to Float64
// in e-notation format and simply ignore other value types, including Uint8, Uint16, Uint32 and Uint64.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) ConvertIntToFloatReading() interfaces.AppFunction {
	conversion := xpert.Conversion{}
	return conversion.ConvertIntToFloatReading
}

// ConvertFloatToIntReading converts readings whose value type is Float32 or Float64 to Int64 by truncating
// the decimal portion and simply ignore other value types.
// This function will return an error and stop the pipeline if a non-edgex event is received or if any error occurs
// during conversion.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) ConvertFloatToIntReading() interfaces.AppFunction {
	conversion := xpert.Conversion{}
	return conversion.ConvertFloatToIntReading
}

// ConvertByteArrayToEvent converts bytes from the previous function to an EdgeX Event.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) ConvertByteArrayToEvent() interfaces.AppFunction {
	conversion := xpert.Conversion{}
	return conversion.ConvertByteArrayToEvent
}

// FilterByValueMaxMin removes readings with outlier value that is out of maximum and minimum as defined in DeviceResource.
// This function only targets on number value, e.g. Int, Float, for non-number readings, the Filter simply ignores.
// This function also only filter when Maximum/Minimum is defined in the DeviceResource; when only Maximum is defined,
// compare with Maximum only; when only Minimum is defined, compare with Minimum only; when both Maximum and Minimum are
// defined, only the readings with number values between Maximum and Minimum will be allowed to pass to next function
// in the pipeline.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) FilterByValueMaxMin() interfaces.AppFunction {
	filter := xpert.Filter{}
	return filter.FilterByValueMaxMin
}

// XpertAddTagsFromDeviceResource adds the configured list of tags of device resources to Events.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) XpertAddTagsFromDeviceResource() interfaces.AppFunction {
	transform := xpert.NewTags()
	return transform.AddTagsFromDeviceResource
}

// XpertAddTagsFromDevice adds the tags from the associated device instance to events.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) XpertAddTagsFromDevice() interfaces.AppFunction {
	transform := xpert.NewTags()
	return transform.AddTagsFromDevice
}

// SetContextVariable sets the context variable per specified parameters.  The name of context variable will be the
// value of parameter VariableName, and the value of context variable will be the value extracted from the specified
// JsonPath of data as passed into the function.  Set continueOnError to true if users would like the pipeline to
// continue even when there is error during SetContextVariable.  When continueOnError is true and error occurs during
// SetContextVariable, the function will return incoming data rather than error. Please note that if incoming data
// doesn't contain specified JsonPath, the context variable will be set to empty string "".
// This function is a configuration function and returns a function pointer.
func (app *Configurable) SetContextVariable(parameters map[string]string) interfaces.AppFunction {

	varName, ok := parameters[VariableName]
	if !ok || len(strings.TrimSpace(varName)) == 0 {
		app.lc.Errorf("SetContextVariable must have mandatory parameter %s specified with non-empty value", VariableName)
		return nil
	}
	variableName := strings.TrimSpace(varName)

	valueJsonPath, ok := parameters[ValueJSONPath]
	if !ok || len(strings.TrimSpace(valueJsonPath)) == 0 {
		app.lc.Errorf("SetContextVariable must have mandatory parameter %s specified with non-empty value", ValueJSONPath)
		return nil
	}
	valueJsonPath = strings.TrimSpace(valueJsonPath)

	var err error
	cor := false
	corVal, ok := parameters[continueOnError]
	if ok {
		cor, err = strconv.ParseBool(corVal)
		if err != nil {
			app.lc.Errorf("unable to parse %s value. error: %s", continueOnError, err)
			return nil
		}
	}

	setter := xpert.NewContextVariableSetter(variableName, valueJsonPath, cor)

	return setter.SetContextVariable
}

func (app *Configurable) processXpertHttpExportParameters(parameters map[string]string) (xpert.HTTPSenderConfig, error) {

	result := xpert.HTTPSenderConfig{}

	method, ok := parameters[ExportMethod]
	if !ok {
		return result, fmt.Errorf("manadatory parameter %s for XpertHTTPExport not found", ExportMethod)
	}
	result.HTTPMethod = strings.ToUpper(method)
	switch result.HTTPMethod {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		// XpertHttpExport supports GET, POST, PUT, PATCH, and DELETE at this moment.
	default:
		return result, fmt.Errorf("unsupported value specified for parameter %s in XpertHTTPExport. Must be value among %s",
			ExportMethod, http.MethodGet+","+http.MethodPost+","+http.MethodPut+","+http.MethodPatch+", or "+http.MethodDelete)
	}

	result.URL, ok = parameters[Url]
	if !ok {
		return result, fmt.Errorf("manadatory parameter %s for XpertHTTPExport not found", Url)
	}
	result.URL = strings.TrimSpace(result.URL)

	result.MimeType, ok = parameters[MimeType]
	if !ok {
		return result, fmt.Errorf("manadatory parameter %s for XpertHTTPExport not found", MimeType)
	}
	result.MimeType = strings.TrimSpace(result.MimeType)

	// PersistOnError is optional and is false by default.
	var value string
	result.PersistOnError = false
	value, ok = parameters[PersistOnError]
	if ok {
		var err error
		result.PersistOnError, err = strconv.ParseBool(value)
		if err != nil {
			return result, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter of XpertHTTPExport. %s", value, PersistOnError, err.Error())
		}
	}

	// ContinueOnSendError is optional and is false by default.
	result.ContinueOnSendError = false
	value, ok = parameters[ContinueOnSendError]
	if ok {
		var err error
		result.ContinueOnSendError, err = strconv.ParseBool(value)
		if err != nil {
			return result, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter of XpertHTTPExport. %s", value, ContinueOnSendError, err.Error())
		}
	}

	// ReturnInputData is optional and is false by default.
	result.ReturnInputData = false
	value, ok = parameters[ReturnInputData]
	if ok {
		var err error
		result.ReturnInputData, err = strconv.ParseBool(value)
		if err != nil {
			return result, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter of XpertHTTPExport. %s", value, ReturnInputData, err.Error())
		}
	}

	// following checks ensure that ContinueOnSendError=true can only be used when ReturnInputData=true and cannot be use when PersistOnError=true
	if result.PersistOnError && result.ContinueOnSendError {
		return result, fmt.Errorf("persistOnError & continueOnSendError can not both be set to true for XpertHTTPExport")
	}
	if result.ContinueOnSendError && !result.ReturnInputData {
		return result, fmt.Errorf("continueOnSendError can only be used in conjunction with returnInputData for multiple XpertHTTPExport")
	}

	// SkipVerify is optional and is false by default.
	result.SkipCertVerify = false
	value, ok = parameters[SkipVerify]
	if ok {
		var err error
		result.SkipCertVerify, err = strconv.ParseBool(value)
		if err != nil {
			return result, fmt.Errorf("could not parse '%s' to a bool for '%s' parameter of XpertHTTPExport. %s", value, SkipVerify, err.Error())
		}
	}

	result.AuthMode = xpert.HTTPAuthModeNONE
	value, ok = parameters[AuthMode]
	if ok {
		result.AuthMode = xpert.HTTPAuthMode(strings.TrimSpace(value))
	}
	value, ok = parameters[SecretPath]
	if ok {
		//even when authMode is none, XpertHTTPExport still needs secretPath to look for cacert secret when users
		//export data to an HTTPS server with self-signed CA cert
		result.SecretPath = strings.TrimSpace(value)
	}
	switch result.AuthMode {
	case xpert.HTTPAuthModeNONE:
	case xpert.HTTPAuthModeHeaderSecret:
		result.HTTPHeaderName = strings.TrimSpace(parameters[HeaderName])
		result.SecretName = strings.TrimSpace(parameters[SecretName])
		if len(result.HTTPHeaderName) == 0 {
			return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s when %s is set to %s", HeaderName, AuthMode, xpert.HTTPAuthModeHeaderSecret)
		}
		if len(result.SecretPath) == 0 {
			return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s when %s is set to %s", SecretPath, AuthMode, xpert.HTTPAuthModeHeaderSecret)
		}
		if len(result.SecretName) == 0 {
			return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s when %s is set to %s", SecretName, AuthMode, xpert.HTTPAuthModeHeaderSecret)
		}
	case xpert.HTTPAuthModeOauth2ClientCredentials:
		if len(result.SecretPath) == 0 {
			return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s when %s is set to %s", SecretPath, AuthMode, xpert.HTTPAuthModeOauth2ClientCredentials)
		}
	case xpert.HTTPAuthModeClientCert:
		if len(result.SecretPath) == 0 {
			return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s when %s is set to %s", SecretPath, AuthMode, result.AuthMode)
		}
		// RenegotiationSupport is optional and is disabled by default.
		result.RenegotiationSupport = tls.RenegotiateNever
		value, ok = parameters[RenegotiationSupport]
		if ok {
			renegotiationValue, err := strconv.ParseInt(value, 10, 0)
			if err != nil {
				return result, fmt.Errorf("could not parse '%s' to a int for '%s' parameter of XpertHTTPExport. %s", value, RenegotiationSupport, err.Error())
			}
			switch renegotiationValue := tls.RenegotiationSupport(renegotiationValue); renegotiationValue {
			case tls.RenegotiateNever, tls.RenegotiateOnceAsClient, tls.RenegotiateFreelyAsClient:
				result.RenegotiationSupport = renegotiationValue
			default:
				return result, fmt.Errorf("unsupported value %d for parameter %s of XpertHTTPExport", renegotiationValue, RenegotiationSupport)
			}
		}
	case xpert.HTTPAuthModeAWSSignature:
		if len(result.SecretPath) == 0 {
			return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s when %s is set to %s", SecretPath, AuthMode, xpert.HTTPAuthModeAWSSignature)
		}
		value, ok = parameters[AWSV4SignerConfigs]
		if ok && len(value) != 0 {
			err := json.Unmarshal([]byte(value), &result.AWSV4SignerConfigs)
			if err != nil {
				return result, fmt.Errorf("could not parse '%s' to a json format for '%s' parameter of XpertHTTPExport. %s", value, AWSV4SignerConfigs, err.Error())
			}
			AWSS4RequiredConfigs := []string{xpert.AWSV4SignerConfigsRegion, xpert.AWSV4SignerConfigsService}
			for _, config := range AWSS4RequiredConfigs {
				if val, ok := result.AWSV4SignerConfigs[config]; !ok || len(val) == 0 {
					return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s in %s when %s is set to %s", config, AWSV4SignerConfigs, AuthMode, xpert.HTTPAuthModeAWSSignature)
				}
			}
		} else {
			return result, fmt.Errorf("XpertHTTPExport missing manadatory parameter %s when %s is set to %s", AWSV4SignerConfigs, AuthMode, xpert.HTTPAuthModeAWSSignature)
		}
	default:
		return result, fmt.Errorf("unsupported AuthMode %s for XpertHTTPExport", result.AuthMode)
	}

	value, ok = parameters[HTTPRequestHeaders]
	if ok && len(value) != 0 {
		err := json.Unmarshal([]byte(value), &result.HTTPRequestHeaders)
		if err != nil {
			return result, fmt.Errorf("could not parse '%s' to a json format for '%s' parameter of XpertHTTPExport. %s", value, HTTPRequestHeaders, err.Error())
		}
	}

	return result, nil
}

// XpertHTTPExport will send data from the previous function to the specified Endpoint via http requests with supported
// method GET, POST, PUT, PATCH, or DELETE and various authentication mechanism. If no previous function exists, the
// event that triggered the pipeline will be used. Passing an empty string to the mimetype method will default to
// application/json.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) XpertHTTPExport(parameters map[string]string) interfaces.AppFunction {
	config, err := app.processXpertHttpExportParameters(parameters)
	if err != nil {
		app.lc.Error(err.Error())
		return nil
	}

	sender := xpert.NewHTTPSender(config)
	return sender.HTTPSend
}

// JavascriptTransform run specified scripts against incoming data.
// It will return an error and stop the pipeline if any error occurs when running specified script or if no data is received.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) JavascriptTransform(parameters map[string]string) interfaces.AppFunction {
	transformScript, ok := parameters[TransformScript]
	if !ok {
		app.lc.Errorf("Could not find mandatory parameter '%s' for JavascriptTransform", TransformScript)
		return nil
	}
	if len(transformScript) == 0 {
		app.lc.Error("TransformScript must be specified with non-empty value")
		return nil
	}

	transform := xpert.NewScriptTransform(transformScript)
	return transform.Transform
}

func (app *Configurable) ConvertDDATAToEvent() interfaces.AppFunction {
	return xpert.NewSparkplugConverter().ConvertDDATAtoEvent
}
