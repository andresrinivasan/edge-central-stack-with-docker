// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
)

const (
	WaitDuration         = 10 * time.Second
	DefaultJWTExpiration = 24 * time.Hour

	BoolTrueNumericValue  = "1"
	BoolFalseNumericValue = "0"

	MQTTTopicPlaceholderStatus    = "{status}"
	MQTTTopicPlaceholderRequestId = "{requestid}"
	MQTTSecretUsername            = messaging.SecretUsernameKey
	MQTTSecretPassword            = messaging.SecretPasswordKey
	MQTTSecretCACert              = messaging.AuthModeCA

	DomainAWS   = "amazonaws.com"
	DomainAzure = "azure-devices.net"

	InfluxDBAuthModeToken          = "token"
	InfluxDBSecretAuthToken        = InfluxDBAuthModeToken
	DefaultInfluxDBMeasurement     = "readings"
	DefaultInfluxDBPrecision       = InfluxDBPrecisionMicroSeconds
	InfluxDBPrecisionNanoSeconds   = "ns"
	InfluxDBPrecisionMicroSeconds  = "us"
	InfluxDBPrecisionMillieSeconds = "ms"
	InfluxDBPrecisionSeconds       = "s"
	InfluxDBValueTypeFloat         = "float"
	InfluxDBValueTypeInteger       = "integer"
	InfluxDBValueTypeUInteger      = "uinteger"
	InfluxDBValueTypeString        = "string"
	InfluxDBValueTypeBoolean       = "boolean"
)
