// Copyright (C) 2022 IOTech Ltd

package app

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
)

const (
	PostgresHost                     = "host"
	PostgresPort                     = "port"
	PostgresDatabaseName             = "databasename"
	PostgresTableName                = "tablename"
	PostgresStoreEventId             = "storeeventid"
	PostgresChunkTimeInterval        = "chunktimeinterval"
	PostgresMaxConn                  = "maxconn"
	DefaultPostgresHost              = "timescaledb"
	DefaultPostgresPort              = "5432"
	DefaultPostgresDatabaseName      = "edgex"
	DefaultPostgresTableName         = "edgex_readings"
	DefaultPostgresChunkTimeInterval = "7 days"
)

// PostgresWrite will write EdgeX readings to PostgreSQL DB
// This function is a configuration function and returns a function pointer.
func (app *Configurable) PostgresWrite(parameters map[string]string) interfaces.AppFunction {
	host, ok := parameters[PostgresHost]
	if !ok {
		app.lc.Warnf("host not found, use default value: %s", DefaultPostgresHost)
		host = DefaultPostgresHost
	}
	port, ok := parameters[PostgresPort]
	if !ok {
		app.lc.Warnf("host not found, use default value: %s", DefaultPostgresPort)
		host = DefaultPostgresPort
	}
	databaseName, ok := parameters[PostgresDatabaseName]
	if !ok {
		app.lc.Warnf("database name not found, use default value: %s", PostgresDatabaseName)
		databaseName = DefaultPostgresDatabaseName
	}
	authMode, ok := parameters[AuthMode]
	if !ok {
		app.lc.Error("mandatory parameter not found:" + AuthMode)
		return nil
	} else {
		authMode = strings.ToLower(authMode)
		if authMode != messaging.AuthModeUsernamePassword {
			app.lc.Errorf("parameter %s specifies unsupported value %s.", AuthMode, authMode)
			return nil
		}
	}
	secretPath, ok := parameters[SecretPath]
	if !ok {
		app.lc.Error("mandatory parameter not found:" + SecretPath)
		return nil
	}
	tableName, ok := parameters[PostgresTableName]
	if !ok {
		app.lc.Warnf("table name not found, use default value: %s", DefaultPostgresTableName)
		tableName = DefaultPostgresTableName
	}
	chunkTimeInterval, ok := parameters[PostgresChunkTimeInterval]
	if !ok {
		app.lc.Warnf("chunk time interval not found, use default value: %s", DefaultPostgresChunkTimeInterval)
		chunkTimeInterval = DefaultPostgresChunkTimeInterval
	}
	maxConn, ok := parameters[PostgresMaxConn]
	if !ok {
		maxConn = fmt.Sprintf("%v", runtime.NumCPU())
		app.lc.Warnf("max conn not found, use default value: %s", maxConn)
	}
	var err error
	storeEventId := false
	storeEventIdVal, ok := parameters[PostgresStoreEventId]
	if ok {
		storeEventId, err = strconv.ParseBool(storeEventIdVal)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to a bool for '%s' parameter", storeEventIdVal, PostgresStoreEventId)
			return nil
		}
	} else {
		app.lc.Infof("%s is not set, use %v as default.", PostgresStoreEventId, storeEventId)
	}
	// PersistOnError is optional and is false by default.
	// If PostgresWrite fails and persistOnError is true and StoreAndForward is enabled,
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

	config := xpert.PostgresWriterConfig{
		Host:              host,
		Port:              port,
		DatabaseName:      databaseName,
		TableName:         tableName,
		StoreEventId:      storeEventId,
		AuthMode:          authMode,
		SecretPath:        secretPath,
		ChunkTimeInterval: chunkTimeInterval,
		MaxConn:           maxConn,
	}
	writer := xpert.NewPostgresWriter(config, persistOnError)
	return writer.PostgresWrite
}
