// Copyright (C) 2021-2023 IOTech Ltd

package xpert

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

var fieldKeyReplaceableRegexp = regexp.MustCompile("{" + common.ResourceName + "}|{" + common.DeviceName + "}|{" + common.ProfileName + "}|{" + common.ValueType + "}")

type InfluxDBWriterConfig struct {
	// ServerURL specifies the InfluxDB Server address
	ServerURL string
	// AuthMode specifies what authentication mode to write points to InfluxDB
	AuthMode string
	// The name of the path in secret provider to retrieve secrets
	SecretPath string
	// Org specifies the InfluxDB Organization
	Org string
	// Bucket specifies the InfluxDB Bucket
	Bucket string
	// Measurement specifies the InfluxDB measurement where EdgeX events shall write to
	Measurement string
	// ValueType specifies what value type of events shall be written into InfluxDB
	// Valid values: float, integer, string, boolean
	ValueType string
	// Precision specifies how much timestamp precision is retained in the InfluxDB points
	// Valid values: ns, us, ms, s
	Precision time.Duration
	// SkipCertVerify specifies whether the Edge Xpert Application Service verifies the server's certificate chain and host name
	SkipCertVerify bool
	// StoreEventTags specifies whether the event tags are exported to InfluxDB
	StoreEventTags bool
	// StoreReadingTags specifies whether the reading tags are exported to InfluxDB
	StoreReadingTags bool
	// FieldKeyPattern specifies the pattern of influxdb field key used to store the reading value
	FieldKeyPattern string
}

type influxDBWriter struct {
	client               influxdb2.Client
	config               InfluxDBWriterConfig
	persistOnError       bool
	secretsLastRetrieved time.Time
	mutex                sync.Mutex
}

type influxDBSecrets struct {
	authToken string
}

func NewInfluxDBWriter(config InfluxDBWriterConfig, persistOnError bool) *influxDBWriter {
	writer := &influxDBWriter{
		client:         nil,
		config:         config,
		persistOnError: persistOnError,
	}
	return writer
}

func (writer *influxDBWriter) getSecrets(ctx interfaces.AppFunctionContext) (*influxDBSecrets, error) {
	secrets, err := ctx.GetSecret(writer.config.SecretPath)
	if err != nil {
		return nil, err
	}
	influxDbSecrets := &influxDBSecrets{
		authToken: secrets[InfluxDBSecretAuthToken],
	}
	return influxDbSecrets, nil
}

func (writer *influxDBWriter) validateSecrets(secrets influxDBSecrets) error {
	switch writer.config.AuthMode {
	case InfluxDBAuthModeToken:
		// need authentication token to make a successful connection
		if len(secrets.authToken) <= 0 {
			return errors.New("mandatory secret authentication token is empty")
		}
	}
	return nil
}

func (writer *influxDBWriter) initializeInfluxDBClient(ctx interfaces.AppFunctionContext) error {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	// If the conditions changed while waiting for the lock, i.e. other thread completed the initialization,
	// then skip doing anything
	if writer.client != nil && !writer.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		return nil
	}

	authToken := ""
	if writer.config.AuthMode != messaging.AuthModeNone {
		//get the secrets from the secret provider and populate the struct
		secrets, err := writer.getSecrets(ctx)
		if err != nil {
			return err
		}
		//ensure that the authmode selected has the required secret values
		if secrets != nil {
			// validate secrets prior to configure TLS
			err := writer.validateSecrets(*secrets)
			if err != nil {
				return err
			}
		}
		authToken = secrets.authToken
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: writer.config.SkipCertVerify, //nolint: gosec
	}
	// Create a client
	writer.client = influxdb2.NewClientWithOptions(writer.config.ServerURL, authToken, influxdb2.DefaultOptions().SetTLSConfig(tlsConfig).SetPrecision(writer.config.Precision))
	// Once producer is successfully initialized, secretsLastRetrieved needs to be updated no matter what the authMode is,
	// as secretsLastRetrieved comparison is part of condition to decide if producer should be initialized.
	writer.secretsLastRetrieved = time.Now()
	return nil
}

func (writer *influxDBWriter) InfluxDBSyncWrite(ctx interfaces.AppFunctionContext, inputData interface{}) (bool, interface{}) {
	if inputData == nil {
		// didn't receive data
		return false, errors.New("no data received")
	}

	var err error = nil
dataTypeDetection:
	switch d := inputData.(type) {
	case dtos.Event: // Scenario for no batch: will receive single event
		err = writer.writeEventsToInfluxDB(ctx, d)
	case []dtos.Event: // Scenario for handle events array, usually happens on resend when Store & Forward is enabled and failures had previously occurred
		err = writer.writeEventsToInfluxDB(ctx, d...)
	case []interface{}: // Scenario for using batch: will receive array of interface{}
		var batchedEvents []dtos.Event
		for _, batchedEvent := range d {
			// InfluxDBSyncWrite only accept batched data of EdgeX Event type.
			// In other words, if there is an InfluxDBSyncWrite function after Batch function, the UseRawDataType parameter of Batch function must be true.
			// Because if UseRawDataType is false, the EdgeX Events to be batched will be marshaled to []byte while batching and must be reconverted here. These conversions are redundant and cause a negative performance impact.
			if be, ok := batchedEvent.(dtos.Event); ok {
				batchedEvents = append(batchedEvents, be)
			} else {
				return false, fmt.Errorf("unsupported data type passed in: %v", reflect.TypeOf(batchedEvent))
			}
		}
		err = writer.writeEventsToInfluxDB(ctx, batchedEvents...)
	case []byte: // Scenario for using Store & Forward mechanism when failures: will receive events converted to []byte no matter whether batch is used or not
		var batchedEvents []dtos.Event
		if err = json.Unmarshal(d, &batchedEvents); err == nil {
			inputData = batchedEvents
			goto dataTypeDetection
		}
		// If the input data is transformed by the JavascriptTransform function, then it may also be an Event.
		var event dtos.Event
		if err = json.Unmarshal(d, &event); err == nil {
			inputData = event
			goto dataTypeDetection
		}
		err = fmt.Errorf("the input data could not be parsed to an EdgeX Event either an array of EdgeX Events")
	default: // receive unsupported data type
		err = fmt.Errorf("unsupported data type passed in: %v", reflect.TypeOf(inputData))
	}
	if err != nil {
		ctx.LoggingClient().Error(err.Error())
		return false, err
	} else {
		return true, inputData
	}
}

func (writer *influxDBWriter) writeEventsToInfluxDB(ctx interfaces.AppFunctionContext, events ...dtos.Event) error {
	// check if client haven't be initialized OR the cache has been invalidated (due to new/updated secrets), need to (re)initialize the client
	if writer.client == nil || writer.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		ctx.LoggingClient().Info("Connecting to InfluxDB server")
		err := writer.initializeInfluxDBClient(ctx)
		if err != nil {
			return err
		}
	}
	writeApi := writer.client.WriteAPIBlocking(writer.config.Org, writer.config.Bucket)

	points, err := writer.toPoints(ctx, events)
	if err != nil {
		return err
	}
	// Write all points at once, and persist events if any writing errors
	if len(points) > 0 {
		if writeError := writeApi.WritePoint(context.Background(), points...); writeError != nil {
			// writing to InfluxDB with errors, convert events to []byte for later retry
			retryData, err := json.Marshal(events)
			if err != nil {
				return fmt.Errorf("failed to marshal the retryData, error: %e", err)
			} else {
				writer.setRetryData(ctx, retryData)
				subMessage := "drop event"
				if writer.persistOnError {
					subMessage = "persisting Events for later retry"
				}
				return fmt.Errorf("failed to write points to influxdb, %s.  Error: %s", subMessage, writeError)
			}
		}
	}
	return nil
}

func (writer *influxDBWriter) setRetryData(ctx interfaces.AppFunctionContext, exportData []byte) {
	if writer.persistOnError {
		ctx.SetRetryData(exportData)
	}
}

// Convert an array of events to an array of influxdb points, note that if fieldKeyPattern doesn't contain any replaceable
// variables as matched to fieldKeyReplaceableRegexp, toPoints will drop those readings whose valueType doesn't match to
// InfluxDBValueType
func (writer *influxDBWriter) toPoints(ctx interfaces.AppFunctionContext, events []dtos.Event) ([]*write.Point, error) {
	// if FieldKeyPattern doesn't contain any replaceable variables as matched to fieldKeyReplaceableRegexp, the
	// fieldKey will be regarded as fixed value and should be processes with valueType check due to the influxdb
	// limitation that the same field of a measurement can only accept the same data type.
	skipValueTypeCheck := fieldKeyReplaceableRegexp.MatchString(writer.config.FieldKeyPattern)

	var points []*write.Point
	// iterate each events
	for _, event := range events {
		//iterate each readings
		for _, reading := range event.Readings {
			// determine if reading's valueType matches to InfluxDBValueType; if yes, ignore this reading
			if !skipValueTypeCheck && !isAcceptedReadingValueType(reading.ValueType, writer.config.ValueType) {
				ctx.LoggingClient().Warnf("reading's valueType %s doesn't match to InfluxDBValueType:%s. drop reading.", reading.ValueType, writer.config.ValueType)
				continue
			}
			point, err := toPoint(reading, writer.config.Measurement, writer.config.FieldKeyPattern,
				writer.config.StoreEventTags, event.Tags, writer.config.StoreReadingTags)
			if err != nil {
				return points, err
			}

			// Use event.Origin if reading.Origin == 0
			origin := reading.Origin
			if origin == 0 {
				origin = event.Origin
			}
			point.SetTime(toNanosecondsTime(origin))

			points = append(points, point)
		}
	}
	return points, nil
}

func toPoint(reading dtos.BaseReading, measurement string, fieldKeyPattern string, storeEventTags bool,
	eventTags map[string]interface{}, storeReadingTags bool) (*write.Point, error) {
	value, err := parseReadingValue(reading)
	// stop pipeline when encountering parsing error
	if err != nil {
		return nil, err
	}
	point := influxdb2.NewPointWithMeasurement(measurement)
	point.AddField(toPointFieldKey(reading, fieldKeyPattern), value)

	// add two mandatory tags for deviceName and resourceName
	point.AddTag(common.DeviceName, reading.DeviceName)
	point.AddTag(common.ResourceName, reading.ResourceName)

	if storeEventTags {
		for k, v := range eventTags {
			point.AddTag(k, fmt.Sprintf("%v", v))
		}
	}

	if storeReadingTags {
		for k, v := range reading.Tags {
			point.AddTag(k, fmt.Sprintf("%v", v))
		}
	}

	return point, nil
}

func toPointFieldKey(reading dtos.BaseReading, fieldKeyPattern string) string {
	attempts := make(map[string]bool)

	result := fieldKeyPattern

	for _, placeholder := range fieldKeyReplaceableRegexp.FindAllString(fieldKeyPattern, -1) {
		if _, tried := attempts[placeholder]; tried {
			continue
		}

		key := strings.TrimRight(strings.TrimLeft(placeholder, "{"), "}")

		replacement := ""
		// replacement only applies to 4 built-in variables: resourceName, profileName, deviceName, and valueType
		switch key {
		case common.ResourceName:
			replacement = reading.ResourceName
		case common.ProfileName:
			replacement = reading.ProfileName
		case common.DeviceName:
			replacement = reading.DeviceName
		case common.ValueType:
			replacement = reading.ValueType
		default:
			// skip the replacements if the key is not in one of the 4 cases
			continue
		}
		result = strings.Replace(result, placeholder, replacement, -1)
		attempts[placeholder] = true
	}

	return result
}

func isAcceptedReadingValueType(readingValueType string, influxDBValueType string) bool {
	switch readingValueType {
	case common.ValueTypeFloat64, common.ValueTypeFloat32:
		return influxDBValueType == InfluxDBValueTypeFloat
	case common.ValueTypeInt8, common.ValueTypeInt16, common.ValueTypeInt32, common.ValueTypeInt64:
		return influxDBValueType == InfluxDBValueTypeInteger
	case common.ValueTypeUint8, common.ValueTypeUint16, common.ValueTypeUint32, common.ValueTypeUint64:
		return influxDBValueType == InfluxDBValueTypeUInteger
	case common.ValueTypeBool:
		return influxDBValueType == InfluxDBValueTypeBoolean
	default:
		return influxDBValueType == InfluxDBValueTypeString
	}
}

func parseReadingValue(reading dtos.BaseReading) (interface{}, error) {
	switch reading.ValueType {
	case common.ValueTypeUint8:
		return parseInteger(reading.Value, 8, false)
	case common.ValueTypeUint16:
		return parseInteger(reading.Value, 16, false)
	case common.ValueTypeUint32:
		return parseInteger(reading.Value, 32, false)
	case common.ValueTypeUint64:
		return parseInteger(reading.Value, 64, false)
	case common.ValueTypeInt8:
		return parseInteger(reading.Value, 8, true)
	case common.ValueTypeInt16:
		return parseInteger(reading.Value, 16, true)
	case common.ValueTypeInt32:
		return parseInteger(reading.Value, 32, true)
	case common.ValueTypeInt64:
		return parseInteger(reading.Value, 64, true)
	case common.ValueTypeFloat32:
		return parseFloat(reading.Value, 32)
	case common.ValueTypeFloat64:
		return parseFloat(reading.Value, 64)
	case common.ValueTypeBool:
		return parseBool(reading.Value)
	case common.ValueTypeBinary:
		return reading.BinaryValue, nil
	default:
		return reading.Value, nil
	}
}

func parseInteger(readingValueStr string, bitSize int, signed bool) (interface{}, error) {
	if len(readingValueStr) == 0 {
		return nil, fmt.Errorf("zero-length reading value found.")
	}
	if signed {
		// parse readingValueStr to integer first, if errors, return nil and error message
		readingValue, err := strconv.ParseInt(readingValueStr, 10, bitSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse non-empty reading value %s to integer with bitSize %v. Error:%s", readingValueStr, bitSize, err)
		}
		return readingValue, nil
	} else {
		// parse readingValueStr to unsigned integer first, if errors, return nil and error message
		readingValue, err := strconv.ParseUint(readingValueStr, 10, bitSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse non-empty reading value %s to unsigned integer with bitSize %v. Error:%s", readingValueStr, bitSize, err)
		}
		return readingValue, nil
	}
}

func parseFloat(readingValueStr string, bitSize int) (interface{}, error) {
	if len(readingValueStr) == 0 {
		return nil, fmt.Errorf("zero-length reading value found")
	}
	floatValue, err := strconv.ParseFloat(readingValueStr, bitSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reading value '%s' to float with bitSize %v. Error:%s", readingValueStr, bitSize, err)
	}
	if bitSize == 32 {
		return float32(floatValue), nil
	}
	return floatValue, nil
}

func parseBool(readingValueStr string) (interface{}, error) {
	if len(readingValueStr) == 0 {
		return nil, fmt.Errorf("zero-length reading value found.")
	}
	readingValue, err := strconv.ParseBool(readingValueStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reading value '%s' to boolean. Error:%s", readingValueStr, err)
	}
	return readingValue, nil
}

// Convert a timestamp to nanoseconds in time.Time type
// If passed timestamp is not in nanoseconds, this function will add trailing zero on the right
// For example, toNanosecondsTime(1581420776817).UnixNano() will return 1581420776817000000
func toNanosecondsTime(timestamp int64) time.Time {
	var offset = len(strconv.FormatInt(time.Now().UnixNano(), 10)) - len(strconv.FormatInt(timestamp, 10))
	if offset == 0 {
		return time.Unix(0, timestamp)
	} else {
		return time.Unix(0, timestamp*int64(math.Pow10(offset)))
	}
}
