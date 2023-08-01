// TODO: Change Copyright to your company if open sourcing or remove header
//
// Copyright (C) 2021 IOTech Ltd

package app

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
)

// TODO: Create your custom configuration functions and remove these samples

// This function is a configuration function and returns a function pointer.
func (app *Configurable) LogEventDetails() interfaces.AppFunction {
	sample := transforms.NewSample()
	return sample.LogEventDetails
}

// This function is a configuration function and returns a function pointer.
func (app *Configurable) ConvertEventToXML() interfaces.AppFunction {
	sample := transforms.NewSample()
	return sample.ConvertEventToXML
}

// This function is a configuration function and returns a function pointer.
func (app *Configurable) OutputXML() interfaces.AppFunction {
	sample := transforms.NewSample()
	return sample.OutputXML
}
