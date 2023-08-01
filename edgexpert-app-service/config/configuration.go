// TODO: Change Copyright to your company if open sourcing or remove header
//
// Copyright (C) 2021 IOTech Ltd

package config

// This file contains example of custom configuration that can be loaded from the service's configuration.toml
// and/or the Configuration Provider, aka Consul (if enabled).
// For more details see https://docs.edgexfoundry.org/2.3/microservices/application/GeneralAppServiceConfig/#custom-configuration
// TODO: Update this configuration as needed for you service's needs and remove this comment
//       or remove this file if not using custom configuration.

import (
	"errors"
	"reflect"
)

// TODO: Define your structured custom configuration types. Must be wrapped with an outer struct with
//
//	single element that matches the top level custom configuration element in your configuration.toml file,
//	'AppCustom' in this example. Replace this example with your configuration structure or
//	remove this file if not using structured custom configuration.
type ServiceConfig struct {
	AppCustom AppCustomConfig
}

// AppCustomConfig is example of service's custom structured configuration that is specified in the service's
// configuration.toml file and Configuration Provider (aka Consul), if enabled.
type AppCustomConfig struct {
	ResourceNames string
	SomeValue     int
	SomeService   HostInfo
}

// HostInfo is example struct for defining connection information for external service
type HostInfo struct {
	Host     string
	Port     int
	Protocol string
}

// TODO: Update using your Custom configuration type.
// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false //errors.New("unable to cast raw config to type 'ServiceConfig'")
	}

	*c = *configuration

	return true
}

// Validate ensures your custom configuration has proper values.
// TODO: Update to properly validate your custom configuration
func (ac *AppCustomConfig) Validate() error {
	if ac.SomeValue <= 0 {
		return errors.New("SomeValue must be greater than zero")
	}

	if reflect.DeepEqual(ac.SomeService, HostInfo{}) {
		return errors.New("SomeService is not set")
	}

	return nil
}
