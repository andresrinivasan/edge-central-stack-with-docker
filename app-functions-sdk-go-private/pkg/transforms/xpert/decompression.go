// Copyright (C) 2022 IOTech Ltd

package xpert

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"io"
)

type Decompression struct {
	Base64Decode bool // indicates the received data should decode with base64
}

// NewDecompression creates, initializes and returns a new instance of Decompression
func NewDecompression() Decompression {
	return Decompression{}
}

// DecompressWithGZIP decode the binary data with base64 encoded string and decompresses from the GZIP algorithm
func (d *Decompression) DecompressWithGZIP(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, fmt.Errorf("function DecompressWithGZIP in pipeline '%s': No Data Received", ctx.PipelineId())
	}
	ctx.LoggingClient().Debugf("Decompression with GZIP in pipeline '%s'", ctx.PipelineId())
	rawData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}
	if d.Base64Decode {
		rawData, err = base64ToBytes(rawData)
		if err != nil {
			return false, errors.NewCommonEdgeX(errors.Kind(err), "unable to decode base64 data", err)
		}
	}
	reader := bytes.NewReader(rawData)
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	ctx.LoggingClient().Debugf("GZIP decompress %v bytes to %v bytes in pipeline '%s'", len(rawData), len(decompressed), ctx.PipelineId())
	return true, decompressed
}

// DecompressWithZLIB decode the binary data with base64 encoded string and decompresses from the ZLIB algorithm
func (d *Decompression) DecompressWithZLIB(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, fmt.Errorf("function DecompressWithZLIB in pipeline '%s': No Data Received", ctx.PipelineId())
	}
	ctx.LoggingClient().Debugf("Decompression with ZLIB in pipeline '%s'", ctx.PipelineId())
	rawData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}
	if d.Base64Decode {
		rawData, err = base64ToBytes(rawData)
		if err != nil {
			return false, errors.NewCommonEdgeX(errors.Kind(err), "unable to decode base64 data", err)
		}
	}
	reader := bytes.NewReader(rawData)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	decompressed, err := io.ReadAll(zlibReader)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	ctx.LoggingClient().Debugf("ZLIB decompress %v bytes to %v bytes in pipeline '%s'", len(rawData), len(decompressed), ctx.PipelineId())
	return true, decompressed
}

func base64ToBytes(data []byte) ([]byte, errors.EdgeX) {
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return decoded, nil
}
