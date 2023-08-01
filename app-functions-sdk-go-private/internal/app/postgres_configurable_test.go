// Copyright (C) 2022 IOTech Ltd

package app

import (
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"

	"github.com/stretchr/testify/assert"
)

func TestConfigurablePostgresWrite(t *testing.T) {
	configurable := Configurable{lc: lc}

	params := make(map[string]string)

	// no mandatory configuration specified in params
	trx := configurable.PostgresWrite(params)
	assert.Nil(t, trx, "return result from PostgresWrite should be nil")

	// no SecretPath
	params[AuthMode] = messaging.AuthModeUsernamePassword
	trx = configurable.PostgresWrite(params)
	assert.Nil(t, trx, "return result from PostgresWrite should be nil")

	// mandatory configuration have been satisfied
	params[SecretPath] = "postgres"
	trx = configurable.PostgresWrite(params)
	assert.NotNil(t, trx, "return result from PostgresWrite should not be nil")

	// mandatory configuration have been satisfied
	params[PostgresTableName] = "foo"
	trx = configurable.PostgresWrite(params)
	assert.NotNil(t, trx, "return result from PostgresWrite should not be nil")

	// unsupported AuthMode
	params[AuthMode] = "unsupported"
	trx = configurable.PostgresWrite(params)
	assert.Nil(t, trx, "return result from PostgresWrite should be nil")
	params[AuthMode] = messaging.AuthModeUsernamePassword

	// unparsable StoreEventId
	params[PostgresStoreEventId] = "abc"
	trx = configurable.PostgresWrite(params)
	assert.Nil(t, trx, "return result from PostgresWrite should be nil")

	// parsable StoreEventId
	params[PostgresStoreEventId] = "true"
	trx = configurable.PostgresWrite(params)
	assert.NotNil(t, trx, "return result from PostgresWrite should not be nil")

	// unparsable persistOnError
	params[PersistOnError] = "ttt"
	trx = configurable.PostgresWrite(params)
	assert.Nil(t, trx, "return result from PostgresWrite should be nil")

	// parsable persistOnError
	params[PersistOnError] = "true"
	trx = configurable.PostgresWrite(params)
	assert.NotNil(t, trx, "return result from PostgresWrite shouldn't be nil")
}
