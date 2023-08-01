// Copyright (C) 2021-2023 IOTech Ltd

package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
)

const (
	InfluxDBServerURL      = "influxdbserverurl"
	InfluxDBOrganization   = "influxdborganization"
	InfluxDBBucket         = "influxdbbucket"
	InfluxDBMeasurement    = "influxdbmeasurement"
	InfluxDBValueType      = "influxdbvaluetype"
	InfluxDBPrecision      = "influxdbprecision"
	StoreEventTags         = "storeeventtags"
	StoreReadingTags       = "storereadingtags"
	FieldKeyPattern        = "fieldkeypattern"
	DefaultFieldKeyPattern = "value"
)

// InfluxDBSyncWrite will convert EdgeX events to influxdb points and then write to target influxdb 2.0
// This function is a configuration function and returns a function pointer.
func (app *Configurable) InfluxDBSyncWrite(parameters map[string]string) interfaces.AppFunction {
	serverURL, ok := parameters[InfluxDBServerURL]
	if !ok {
		app.lc.Error("mandatory parameter not found:" + InfluxDBServerURL)
		return nil
	}
	authMode, ok := parameters[AuthMode]
	if ok {
		authMode = strings.ToLower(authMode)
		// Check if supported authMode is specified; influxdb only supports token authentication as of 2020/07/15
		if authMode != xpert.InfluxDBAuthModeToken && authMode != messaging.AuthModeNone {
			app.lc.Errorf("parameter %s specifies unsupported value %s.", AuthMode, authMode)
			return nil
		}
	} else {
		app.lc.Debugf("%s parameter is not set, use %s as default", AuthMode, messaging.AuthModeNone)
		authMode = messaging.AuthModeNone
	}
	//default secretPath to empty string, and expects to have non-empty value when authMode is not none
	secretPath := ""
	if authMode != messaging.AuthModeNone {
		secretPath, ok = parameters[SecretPath]
		if !ok {
			app.lc.Error("mandatory parameter not found:" + SecretPath)
			return nil
		}
	}
	org, ok := parameters[InfluxDBOrganization]
	if !ok {
		app.lc.Warnf("%s parameter is not set, use empty string as default", InfluxDBOrganization)
		org = ""
	}
	bucket, ok := parameters[InfluxDBBucket]
	if !ok {
		app.lc.Warnf("%s parameter is not set, use empty string as default", InfluxDBBucket)
		bucket = ""
	}
	measurement, ok := parameters[InfluxDBMeasurement]
	if !ok {
		app.lc.Warnf("%s parameter is not set, use %v as default", InfluxDBMeasurement, xpert.DefaultInfluxDBMeasurement)
		measurement = xpert.DefaultInfluxDBMeasurement
	}
	valueType := xpert.InfluxDBValueTypeFloat
	valueTypeStr, ok := parameters[InfluxDBValueType]
	if ok {
		valueTypeStr = strings.ToLower(valueTypeStr)
		switch valueTypeStr {
		case xpert.InfluxDBValueTypeFloat, xpert.InfluxDBValueTypeInteger, xpert.InfluxDBValueTypeUInteger, xpert.InfluxDBValueTypeString, xpert.InfluxDBValueTypeBoolean:
			valueType = valueTypeStr
		default:
			app.lc.Errorf("unrecognized influxdb value type %s.", valueTypeStr)
			return nil
		}
	} else {
		app.lc.Infof("%s parameter is not set, use %v as default", InfluxDBValueType, valueType)
	}
	precision := time.Microsecond
	precisionStr, ok := parameters[InfluxDBPrecision]
	if ok {
		precisionStr = strings.ToLower(precisionStr)
		switch precisionStr {
		case xpert.InfluxDBPrecisionNanoSeconds:
			precision = time.Nanosecond
		case xpert.InfluxDBPrecisionMicroSeconds:
			precision = time.Microsecond
		case xpert.InfluxDBPrecisionMillieSeconds:
			precision = time.Millisecond
		case xpert.InfluxDBPrecisionSeconds:
			precision = time.Second
		default:
			app.lc.Errorf("unrecognized precision %s.", precisionStr)
			return nil
		}
	} else {
		app.lc.Warnf("%s parameter is not set, use %v as default", InfluxDBPrecision, xpert.DefaultInfluxDBPrecision)
	}
	var err error
	skipCertVerify := false
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to a bool for '%s' parameter", skipVerifyVal, SkipVerify)
			return nil
		}
	} else {
		app.lc.Infof("%s is not set, use %v as default.", SkipVerify, skipCertVerify)
	}
	// PersistOnError is optional and is false by default.
	// If the Kafka send fails and persistOnError is true and StoreAndForward is enabled,
	// the data will be stored for later retry.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to a bool for '%s' parameter", value, PersistOnError)
			return nil
		}
	}

	storeEventTags := false
	storeEventTagsVal, ok := parameters[StoreEventTags]
	if ok {
		storeEventTags, err = strconv.ParseBool(storeEventTagsVal)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to a bool for '%s' parameter", storeEventTagsVal, StoreEventTags)
			return nil
		}
	}

	storeReadingTags := false
	storeReadingTagsVal, ok := parameters[StoreReadingTags]
	if ok {
		storeReadingTags, err = strconv.ParseBool(storeReadingTagsVal)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to a bool for '%s' parameter", storeReadingTagsVal, StoreReadingTags)
			return nil
		}
	}

	fieldKeyPattern, ok := parameters[FieldKeyPattern]
	if !ok {
		fieldKeyPattern = DefaultFieldKeyPattern
		app.lc.Infof("%s parameter is not set, use %v as default", FieldKeyPattern, fieldKeyPattern)
	}

	config := xpert.InfluxDBWriterConfig{
		ServerURL:        serverURL,
		AuthMode:         authMode,
		SecretPath:       secretPath,
		Org:              org,
		Bucket:           bucket,
		Measurement:      measurement,
		ValueType:        valueType,
		Precision:        precision,
		SkipCertVerify:   skipCertVerify,
		StoreEventTags:   storeEventTags,
		StoreReadingTags: storeReadingTags,
		FieldKeyPattern:  fieldKeyPattern,
	}
	writer := xpert.NewInfluxDBWriter(config, persistOnError)
	return writer.InfluxDBSyncWrite
}
