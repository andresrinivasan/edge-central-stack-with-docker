// Copyright (C) 2021 IOTech Ltd

package app

import (
	"strconv"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

const (
	deviceNameJSONPath    = "devicenamejsonpath"
	commandNameJSONPath   = "commandnamejsonpath"
	bodyJSONPath          = "bodyjsonpath"
	requestMethodJSONPath = "requestmethodjsonpath"
	defaultRequestMethod  = "defaultrequestmethod"
	pushEvent             = "pushevent"
	returnEvent           = "returnevent"
	continueOnError       = "continueonerror"
)

// ExecuteCoreCommand parses the incoming JSON data to determine the core command and then issues it through the CommandClient.
// When configuring functions pipeline, an export function can be set after this function to send the command execution result.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) ExecuteCoreCommand(parameters map[string]string) interfaces.AppFunction {
	d, ok := parameters[deviceNameJSONPath]
	if !ok {
		app.lc.Errorf("mandatory parameter %s not found.", deviceNameJSONPath)
		return nil
	}

	c, ok := parameters[commandNameJSONPath]
	if !ok {
		app.lc.Errorf("mandatory parameter %s not found.", commandNameJSONPath)
		return nil
	}

	dm := parameters[defaultRequestMethod]
	m := parameters[requestMethodJSONPath]
	if len(dm) == 0 && len(m) == 0 {
		app.lc.Errorf("one of %s and %s must be specified.", defaultRequestMethod, requestMethodJSONPath)
		return nil
	}

	b := parameters[bodyJSONPath]

	// PushEvent is optional, and "no" is used by default.
	// If set to yes, a successful GET will result in an event being pushed to the EdgeX system.
	p, ok := parameters[pushEvent]
	if ok {
		if p != common.ValueYes && p != common.ValueNo {
			app.lc.Errorf("invalid %s value: %s, should be %s or %s", pushEvent, p, common.ValueYes, common.ValueNo)
			return nil
		}
	} else {
		app.lc.Infof("%s is not set, use %v as default.", pushEvent, common.ValueNo)
		p = common.ValueNo
	}

	// ReturnEvent is optional, and "yes" is used by default.
	// If set to no, a successful GET will not return an Event.
	r, ok := parameters[returnEvent]
	if ok {
		if r != common.ValueYes && r != common.ValueNo {
			app.lc.Errorf("invalid %s value: %s, should be %s or %s", returnEvent, r, common.ValueYes, common.ValueNo)
			return nil
		}
	} else {
		app.lc.Infof("%s is not set, use %v as default.", returnEvent, common.ValueYes)
		r = common.ValueYes
	}

	var err error
	cor := false
	// continueOnError is useful for chained export functions. If true, the functions pipeline continues after an error
	// occurs so next export function executes.
	corVal, ok := parameters[continueOnError]
	if ok {
		cor, err = strconv.ParseBool(corVal)
		if err != nil {
			app.lc.Errorf("unable to parse %s value. error: %s", continueOnError, err)
			return nil
		}
	}

	commandExecutor := xpert.CoreCommandExecutor{
		DeviceNameJSONPath:    d,
		CommandNameJSONPath:   c,
		BodyJSONPath:          b,
		RequestMethodJSONPath: m,
		DefaultRequestMethod:  dm,
		PushEvent:             p,
		ReturnEvent:           r,
		ContinueOnError:       cor}

	return commandExecutor.Execute
}
