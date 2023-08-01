// Copyright (C) 2021 IOTech Ltd

package app

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigurableKafkaSend(t *testing.T) {
	configurable := Configurable{lc: lc}

	params := make(map[string]string)

	// no mandatory configuration specified in params
	trx := configurable.KafkaSend(params)
	assert.Nil(t, trx, "return result from KafkaSend should be nil")

	// no Address, Port, and Topic
	params[KafkaClientID] = "edgex"
	trx = configurable.KafkaSend(params)
	assert.Nil(t, trx, "return result from KafkaSend should be nil")

	// no Port, and Topic
	params[KafkaAddress] = "localhost"
	trx = configurable.KafkaSend(params)
	assert.Nil(t, trx, "return result from KafkaSend should be nil")

	// no Topic
	params[KafkaPort] = "9092"
	trx = configurable.KafkaSend(params)
	assert.Nil(t, trx, "return result from KafkaSend should be nil")

	// mandatory configuration have been satisfied
	params[KafkaTopic] = "test"
	trx = configurable.KafkaSend(params)
	assert.NotNil(t, trx, "return result from KafkaSend should not be nil")

	// unparsible skipCertVerify
	params[SkipVerify] = "unparsible"
	trx = configurable.KafkaSend(params)
	assert.Nil(t, trx, "return result from KafkaSend should be nil")

	// parsible skipCertVerify
	params[SkipVerify] = "true"
	trx = configurable.KafkaSend(params)
	assert.NotNil(t, trx, "return result from KafkaSend should not be nil")

	// unsupported authmode
	params[AuthMode] = "unsupported"
	trx = configurable.KafkaSend(params)
	assert.Nil(t, trx, "return result from KafkaSend should be nil")

	// supported authmode none
	params[AuthMode] = messaging.AuthModeNone
	trx = configurable.KafkaSend(params)
	assert.NotNil(t, trx, "return result from KafkaSend should not be nil")

	// supported authmode clientcert but no secretpath
	params[AuthMode] = messaging.AuthModeCert
	trx = configurable.KafkaSend(params)
	assert.Nil(t, trx, "return result from KafkaSend should be nil")

	// supported authmode clientcert with secretpath
	params[SecretPath] = "kafka"
	trx = configurable.KafkaSend(params)
	assert.NotNil(t, trx, "return result from KafkaSend should not be nil")
}
