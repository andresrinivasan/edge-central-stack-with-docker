// Copyright (C) 2022-2023 IOTech Ltd

package xpert

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cast"
)

const (
	PostgresTableValueType               = "edgex_valuetypes"
	PostgresTableResource                = "edgex_resources"
	PostgresPlaceholderTable             = "$table"
	PostgresPlaceholderColumnValue       = "$column"
	PostgresAuthModeUsernamePassword     = messaging.AuthModeUsernamePassword
	PostgresColumnId                     = "id"
	PostgresColumnTimestamp              = "timestamp"
	PostgresColumnDeviceName             = "device_name"
	PostgresColumnResourceName           = "resource_name"
	PostgresColumnResourceId             = "resource_id"
	PostgresColumnTags                   = "tags"
	PostgresColumnValueType              = "value_type"
	PostgresColumnValueTypeId            = "value_type_id"
	PostgresColumnEventId                = "event_id"
	PostgresErrUniqueConstraintViolation = "SQLSTATE 23505"
	// SQL statements ...
	PostgresInsertReadingWithEventId = "INSERT INTO " + PostgresPlaceholderTable +
		"(" + PostgresPlaceholderColumnValue + ", " + PostgresColumnTimestamp + ", " + PostgresColumnResourceId + ", " +
		PostgresColumnValueTypeId + ", " + PostgresColumnEventId + ") " + "VALUES ($1, $2, $3, $4, $5);"
	PostgresInsertReadingWithoutEventId = "INSERT INTO " + PostgresPlaceholderTable +
		"(" + PostgresPlaceholderColumnValue + ", " + PostgresColumnTimestamp + ", " + PostgresColumnResourceId + ", " +
		PostgresColumnValueTypeId + ") " + "VALUES ($1, $2, $3, $4);"
	PostgresInsertValueType = "INSERT INTO " + PostgresTableValueType + " (" + PostgresColumnId + ", " +
		PostgresColumnValueType + ") " + "SELECT $1, $2 WHERE NOT EXISTS ( SELECT 1 FROM " + PostgresTableValueType +
		" WHERE " + PostgresColumnId + "=$1 AND " + PostgresColumnValueType + "=$2)"
	PostgresInsertResource = "INSERT INTO " + PostgresTableResource + "(" + PostgresColumnDeviceName + "," +
		PostgresColumnResourceName + "," + PostgresColumnTags + ") " +
		"SELECT $1, $2, $3 WHERE NOT EXISTS ( SELECT 1 FROM " + PostgresTableResource +
		" WHERE " + PostgresColumnDeviceName + " = $1 AND " + PostgresColumnResourceName + " = $2 AND " +
		PostgresColumnTags + " =$3 ) RETURNING " + PostgresColumnId
	PostgresQueryResourceId = "SELECT " + PostgresColumnId + " FROM " + PostgresTableResource + " WHERE " +
		PostgresColumnDeviceName + "=$1 AND " + PostgresColumnResourceName + "=$2 AND " + PostgresColumnTags + "=$3"
	PostgresQueryValueTypes      = "SELECT * FROM " + PostgresTableValueType
	PostgresQueryResources       = "SELECT * FROM " + PostgresTableResource
	PostgresCreateTableValueType = "CREATE TABLE IF NOT EXISTS " + PostgresTableValueType + " (" +
		PostgresColumnId + " SMALLINT," +
		PostgresColumnValueType + " TEXT NOT NULL," +
		"PRIMARY KEY (" + PostgresColumnId + "))"
	PostgresCreateTableResource = "CREATE TABLE IF NOT EXISTS " + PostgresTableResource + " (" +
		PostgresColumnId + " SERIAL," +
		PostgresColumnDeviceName + " TEXT NOT NULL," +
		PostgresColumnResourceName + " TEXT NOT NULL," +
		PostgresColumnTags + " JSONB," +
		"PRIMARY KEY (" + PostgresColumnId + "));"
	PostgresCreateEdgexResourcesIndex1 = "CREATE UNIQUE INDEX IF NOT EXISTS edgex_resources_index1 ON edgex_resources (device_name, resource_name, tags) WHERE tags IS NOT NULL;"
	PostgresCreateEdgexResourcesIndex2 = "CREATE UNIQUE INDEX IF NOT EXISTS edgex_resources_index2 ON edgex_resources (device_name, resource_name) WHERE tags IS NULL;"

	PostgresColumnValue        = "value"
	PostgresCreateTableReading = "CREATE TABLE IF NOT EXISTS " + PostgresPlaceholderTable + " (" +
		PostgresColumnTimestamp + " TIMESTAMP," +
		PostgresColumnResourceId + " INTEGER," +
		PostgresColumnValueTypeId + " SMALLINT," +
		PostgresColumnEventId + " UUID," +
		PostgresColumnValue + " BYTEA )"
)

type PostgresWriterConfig struct {
	Host              string
	Port              string
	DatabaseName      string
	TableName         string
	StoreEventId      bool
	SecretPath        string
	AuthMode          string
	ChunkTimeInterval string
	MaxConn           string
}

type postgresWriter struct {
	conn                 *pgxpool.Pool
	config               PostgresWriterConfig
	persistOnError       bool
	secretsLastRetrieved time.Time
	mutex                sync.Mutex
}

type postgresSecrets struct {
	username string
	password string
}

type sqlArguments struct {
	value       []byte
	timestamp   interface{}
	resourceId  interface{}
	valueTypeId interface{}
	eventId     interface{}
}

func (sa sqlArguments) toSlice() []interface{} {
	// For the order of arguments, see PostgresInsertReadingWithEventId and PostgresInsertReadingWithoutEventId const
	s := []interface{}{sa.value, sa.timestamp, sa.resourceId, sa.valueTypeId}
	if sa.eventId != nil {
		s = append(s, sa.eventId)
	}
	return s
}

func NewPostgresWriter(config PostgresWriterConfig, persistOnError bool) *postgresWriter {
	writer := &postgresWriter{
		conn:           nil,
		config:         config,
		persistOnError: persistOnError,
	}
	return writer
}

func (w *postgresWriter) getSecrets(ctx interfaces.AppFunctionContext) (*postgresSecrets, error) {
	secrets, err := ctx.GetSecret(w.config.SecretPath)
	if err != nil {
		return nil, err
	}
	postgresSecrets := &postgresSecrets{
		username: secrets[messaging.SecretUsernameKey],
		password: secrets[messaging.SecretPasswordKey],
	}
	w.secretsLastRetrieved = time.Now()
	return postgresSecrets, nil
}

func (w *postgresWriter) validateSecrets(secrets postgresSecrets) error {
	switch w.config.AuthMode {
	case PostgresAuthModeUsernamePassword:
		if len(secrets.username) == 0 {
			return fmt.Errorf("auth mode %s selected however %s was not found at secret path",
				PostgresAuthModeUsernamePassword, messaging.SecretUsernameKey)
		}
		if len(secrets.password) == 0 {
			return fmt.Errorf("auth mode %s selected however %s was not found at secret path",
				PostgresAuthModeUsernamePassword, messaging.SecretPasswordKey)
		}
	default:
		return fmt.Errorf("invalid authentication mode: %s", w.config.AuthMode)
	}
	return nil
}

func (w *postgresWriter) initializePostgresClient(ctx interfaces.AppFunctionContext) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// If the conditions changed while waiting for the lock, i.e. other thread completed the initialization,
	// then skip doing anything
	if w.conn != nil && !w.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		return nil
	}

	// Get the secrets from the secret provider and populate the struct
	secrets, err := w.getSecrets(ctx)
	if err != nil {
		return err
	} else {
		err := w.validateSecrets(*secrets)
		if err != nil {
			return err
		}
	}

	// Create a connection pool
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=%s",
		secrets.username, secrets.password, w.config.Host, w.config.Port, w.config.DatabaseName, w.config.MaxConn)
	conn, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to database, error: %s", err)
	}
	ctx.LoggingClient().Infof("Connected the Postgres with config: %+v.", conn.Config())

	// Create relational table edgex_valuetypes
	_, err = conn.Exec(context.Background(), PostgresCreateTableValueType)
	if err != nil {
		return fmt.Errorf("failed to create table \"%s\", error: %s", PostgresTableValueType, err)
	}

	// Insert value types
	batch := &pgx.Batch{}
	for id, valueType := range cache.HardcodedValueTypes {
		batch.Queue(PostgresInsertValueType, id, valueType)
	}
	br := conn.SendBatch(context.Background(), batch)
	if _, err := br.Exec(); err != nil {
		return fmt.Errorf("failed to insert value type name into %s tabe, err: %v", PostgresTableValueType, err)
	}
	err = br.Close()
	if err != nil {
		return fmt.Errorf("failed to closes the batch operation and release the conn to conn pool, err: %v", err)
	}

	// Create relational table edgex_resources
	_, err = conn.Exec(context.Background(), PostgresCreateTableResource)
	if err != nil {
		return fmt.Errorf("failed to create table \"%s\", error: %s", PostgresTableResource, err)
	}

	_, err = conn.Exec(context.Background(), PostgresCreateEdgexResourcesIndex1)
	if err != nil {
		return fmt.Errorf("failed to create index on table \"%s\", error: %s", PostgresTableResource, err)
	}
	_, err = conn.Exec(context.Background(), PostgresCreateEdgexResourcesIndex2)
	if err != nil {
		return fmt.Errorf("failed to create index on table \"%s\", error: %s", PostgresTableResource, err)
	}

	// Create main table
	sqlCreateTable := strings.ReplaceAll(PostgresCreateTableReading, PostgresPlaceholderTable, w.config.TableName)
	_, err = conn.Exec(context.Background(), sqlCreateTable)
	if err != nil {
		return fmt.Errorf("failed to create table \"%s\", error: %s", w.config.TableName, err)
	}

	// Call TimescaleDB SQL functions to create hypertable and set chunk time interval
	// https://docs.timescale.com/api/latest#timescaledb-api-reference
	sqlCreateHyperTable := fmt.Sprintf(`SELECT create_hypertable('%s', '%s', if_not_exists => TRUE);`,
		w.config.TableName, PostgresColumnTimestamp)
	_, err = conn.Exec(context.Background(), sqlCreateHyperTable)
	if err != nil {
		ctx.LoggingClient().Warnf("Unable to create hypertable, error: %s. Data will be stored in standard table.", err)
	} else {
		// Set chunk time interval
		sqlSetChunkTimeInterval := fmt.Sprintf(`SELECT set_chunk_time_interval('%s', INTERVAL '%s');`,
			w.config.TableName, w.config.ChunkTimeInterval)
		_, err = conn.Exec(context.Background(), sqlSetChunkTimeInterval)
		if err != nil {
			ctx.LoggingClient().Warnf("Unable to set chunk time interval, error: %s.", err)
		}
	}

	w.conn = conn
	return nil
}

func (w *postgresWriter) PostgresWrite(ctx interfaces.AppFunctionContext, inputData interface{}) (bool, interface{}) {
	if inputData == nil {
		// didn't receive data
		return false, errors.New("no data received")
	}

	var err error = nil
dataTypeDetection:
	switch d := inputData.(type) {
	case dtos.Event: // Scenario for no batch: will receive single event
		err = w.writeReadingsToPostgres(ctx, d)
	case []dtos.Event: // Scenario for handle events array, usually happens on resend when Store & Forward is enabled and failures had previously occurred
		err = w.writeReadingsToPostgres(ctx, d...)
	case []interface{}: // Scenario for using batch: will receive array of interface{}
		var batchedEvents []dtos.Event
		for _, batchedEvent := range d {
			// If there is a PostgresWrite function after Batch function, the UseRawDataType parameter of Batch function must be true.
			// Because if UseRawDataType is false, the EdgeX Events to be batched will be marshaled to []byte while batching and must be reconverted here. These conversions are redundant and cause a negative performance impact.
			if be, ok := batchedEvent.(dtos.Event); ok {
				batchedEvents = append(batchedEvents, be)
			} else {
				return false, fmt.Errorf("unsupported data type passed in: %v", reflect.TypeOf(batchedEvent))
			}
		}
		err = w.writeReadingsToPostgres(ctx, batchedEvents...)
	case []byte: // Scenario for using Store & Forward mechanism when failures: will receive events converted to []byte no matter whether batch is used or not
		var batchedEvents []dtos.Event
		if err = json.Unmarshal(d, &batchedEvents); err == nil {
			inputData = batchedEvents
			goto dataTypeDetection
		}
	default: // receive unsupported data type
		err = fmt.Errorf("unsupported data type passed in: %v", reflect.TypeOf(inputData))
	}
	if err != nil {
		ctx.LoggingClient().Error(err.Error())
		return false, err
	} else {
		return true, nil
	}
}

func (w *postgresWriter) writeReadingsToPostgres(ctx interfaces.AppFunctionContext, events ...dtos.Event) error {
	// check if client haven't been initialized OR the cache has been invalidated (due to new/updated secrets), need to (re)initialize the client
	if w.conn == nil || w.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		err := w.initializePostgresClient(ctx)
		if err != nil {
			return err
		}
	}

	var setRetryData bool
	conn, err := w.conn.Acquire(context.Background())
	if err == nil {
		defer conn.Release()
		ch := cacheHelper{ctx, conn}
		insertionErrors := make(map[string]error)

		//create batch
		batch := &pgx.Batch{}
		for _, e := range events {
			for _, r := range e.Readings {
				sql, args, err := getSqlStatementAndArguments(ch, w.config, r, e.Id, e.Tags)
				if err != nil {
					insertionErrors[r.Id] = err
					continue
				}
				batch.Queue(sql, args.toSlice()...)
			}
		}
		br := conn.SendBatch(context.Background(), batch)
		defer br.Close()
		if _, err := br.Exec(); err != nil {
			ctx.LoggingClient().Errorf("failed to insert readings into Postgres, err: %v", err)
			setRetryData = true
		}
	} else {
		ctx.LoggingClient().Errorf("failed to acquire a Postgres connection from the pool, error: %v", err)
		setRetryData = true
	}

	if setRetryData {
		// writing to Postgres with errors, convert events to []byte for later retry
		retryData, err := json.Marshal(events)
		if err != nil {
			return fmt.Errorf("failed to marshal the retryData, error: %e", err)
		} else {
			w.setRetryData(ctx, retryData)
			subMessage := "drop event"
			if w.persistOnError {
				subMessage = "persisting Events for later retry"
			}
			return fmt.Errorf("failed to write readings to Postgres, %s", subMessage)
		}
	}

	return nil
}

func (w *postgresWriter) setRetryData(ctx interfaces.AppFunctionContext, exportData []byte) {
	if w.persistOnError {
		ctx.SetRetryData(exportData)
	}
}

func getSqlStatementAndArguments(cacheHelper PostgresCacheHelper, config PostgresWriterConfig,
	reading dtos.BaseReading, eventId string, tags map[string]interface{}) (string, sqlArguments, error) {

	resourceId, err := cacheHelper.GetResourceId(reading, tags)
	if err != nil {
		return "", sqlArguments{}, fmt.Errorf("failed to get resource ID from cache, err: %v", err)
	}

	valueTypeId, err := cacheHelper.GetValueTypeId(reading)
	if err != nil {
		return "", sqlArguments{}, fmt.Errorf("failed to get value type ID from cache, err: %v", err)
	}

	arguments := sqlArguments{
		timestamp:   toNanosecondsTime(reading.Origin),
		resourceId:  resourceId,
		valueTypeId: valueTypeId,
	}
	switch reading.ValueType {
	case common.ValueTypeBinary:
		arguments.value = reading.BinaryValue
	case common.ValueTypeObject, common.ValueTypeObjectArray:
		jsonEncodedData, err := json.Marshal(reading.ObjectValue)
		if err != nil {
			return "", sqlArguments{}, fmt.Errorf("failed to encode obejct value, err: %v", err)
		}
		arguments.value = jsonEncodedData
	default:
		binaryData, err := convertReadingToBinaryData(reading)
		if err != nil {
			return "", sqlArguments{}, fmt.Errorf("failed to convert reading to binary data, %v", err)
		}
		arguments.value = binaryData
	}

	sql := PostgresInsertReadingWithoutEventId
	if config.StoreEventId {
		sql = PostgresInsertReadingWithEventId
		arguments.eventId = eventId
	}

	sql = strings.ReplaceAll(
		strings.ReplaceAll(sql, PostgresPlaceholderTable, config.TableName),
		PostgresPlaceholderColumnValue, PostgresColumnValue)

	return sql, arguments, nil
}

type PostgresCacheHelper interface {
	GetResourceId(reading dtos.BaseReading, tags map[string]interface{}) (int32, error)
	GetValueTypeId(reading dtos.BaseReading) (int16, error)
}

type cacheHelper struct {
	appCtx  interfaces.AppFunctionContext
	pgxConn *pgxpool.Conn
}

func (ch cacheHelper) GetResourceId(reading dtos.BaseReading, tags map[string]interface{}) (int32, error) {
	var resourceId int32
	resourceCache := cache.PostgresDeviceResources(ch.appCtx, ch.pgxConn, PostgresQueryResources)
	resource, ok := resourceCache.Get(reading.DeviceName, reading.ResourceName, tags)
	if !ok {
		err := ch.pgxConn.QueryRow(context.Background(), PostgresInsertResource, reading.DeviceName, reading.ResourceName, tags).Scan(&resourceId)
		switch err {
		case pgx.ErrNoRows: // data has been inserted by other goroutines
			// self-call to get resource id from the cache
			return ch.GetResourceId(reading, tags)
		case nil:
			if ok := resourceCache.Add(resourceId, reading, tags); !ok {
				return resourceId, errors.New("failed to add resource data to cache")
			}
		default:
			if strings.Contains(err.Error(), PostgresErrUniqueConstraintViolation) {
				// SQLSTATE 23505 returned when a unique constraint is violated
				// self-call to get resource id from the cache
				return ch.GetResourceId(reading, tags)
			}
			return resourceId, fmt.Errorf("failed to insert data into the table %s, err: %v", PostgresTableResource, err)
		}
	} else {
		resourceId = resource.Id
	}
	return resourceId, nil
}

func (ch cacheHelper) GetValueTypeId(reading dtos.BaseReading) (int16, error) {
	var valueTypeId int16
	valueTypeCache := cache.PostgresValueTypes(ch.appCtx, ch.pgxConn, PostgresQueryValueTypes)
	valueType, ok := valueTypeCache.Get(reading.ValueType)
	if !ok {
		return valueTypeId, fmt.Errorf("invalid value type: %s", reading.ValueType)
	} else {
		valueTypeId = valueType.Id
	}
	return valueTypeId, nil
}

func convertReadingToBinaryData(reading dtos.BaseReading) ([]byte, error) {
	var value interface{}
	var err error
	switch reading.ValueType {
	case common.ValueTypeBool:
		value, err = cast.ToBoolE(reading.Value)
	case common.ValueTypeInt8:
		value, err = cast.ToInt8E(reading.Value)
	case common.ValueTypeUint8:
		value, err = cast.ToUint8E(reading.Value)
	case common.ValueTypeInt16:
		value, err = cast.ToInt16E(reading.Value)
	case common.ValueTypeUint16:
		value, err = cast.ToUint16E(reading.Value)
	case common.ValueTypeInt32:
		value, err = cast.ToInt32E(reading.Value)
	case common.ValueTypeUint32:
		value, err = cast.ToUint32E(reading.Value)
	case common.ValueTypeInt64:
		value, err = cast.ToInt64E(reading.Value)
	case common.ValueTypeUint64:
		value, err = cast.ToUint64E(reading.Value)
	case common.ValueTypeFloat32:
		value, err = cast.ToFloat32E(reading.Value)
	case common.ValueTypeFloat64:
		value, err = cast.ToFloat64E(reading.Value)
	case common.ValueTypeString:
		return []byte(reading.Value), nil // cast the string to byte array because binary.Write doesn't handle string
	case common.ValueTypeBoolArray, common.ValueTypeStringArray, common.ValueTypeInt8Array, common.ValueTypeUint8Array,
		common.ValueTypeInt16Array, common.ValueTypeUint16Array, common.ValueTypeInt32Array, common.ValueTypeUint32Array,
		common.ValueTypeInt64Array, common.ValueTypeUint64Array, common.ValueTypeFloat32Array, common.ValueTypeFloat64Array:
		return []byte(reading.Value), nil // let user parse array from the string value
	default:
		return nil, fmt.Errorf("unsupported value type: %s", reading.ValueType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to convert reading string value to specified data type %s: %v", reading.ValueType, err)
	}
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert reading value to binary: %v", err)
	}
	return buf.Bytes(), nil
}
