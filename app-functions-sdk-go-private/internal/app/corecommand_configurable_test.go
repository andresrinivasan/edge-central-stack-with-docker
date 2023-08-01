package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigurableExecuteCoreCommand(t *testing.T) {
	configurable := Configurable{lc: lc}
	params := make(map[string]string)
	msgResultShouldBeNil := "return result from ExecuteCoreCommand should be nil"
	msgResultShouldNotBeNil := "return result from ExecuteCoreCommand shouldn't be nil"

	// no mandatory configuration specified in params
	trx := configurable.ExecuteCoreCommand(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// device name JSON path
	params[deviceNameJSONPath] = "device"
	trx = configurable.ExecuteCoreCommand(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// device name JSON path and command name JSON path
	params[commandNameJSONPath] = "command"
	trx = configurable.ExecuteCoreCommand(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// device name JSON path, command name JSON path, and request method; all satisfied
	params[requestMethodJSONPath] = "method"
	trx = configurable.ExecuteCoreCommand(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable PushEvent
	params[pushEvent] = "true"
	trx = configurable.ExecuteCoreCommand(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable PushEvent
	params[pushEvent] = "yes"
	trx = configurable.ExecuteCoreCommand(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable ReturnEvent
	params[returnEvent] = "abc"
	trx = configurable.ExecuteCoreCommand(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable ReturnEvent
	params[returnEvent] = "yes"
	trx = configurable.ExecuteCoreCommand(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

	// unparsable ContinueOnError
	params[continueOnError] = "abc"
	trx = configurable.ExecuteCoreCommand(params)
	assert.Nil(t, trx, msgResultShouldBeNil)

	// parsable ContinueOnError
	params[continueOnError] = "true"
	trx = configurable.ExecuteCoreCommand(params)
	assert.NotNil(t, trx, msgResultShouldNotBeNil)

}
