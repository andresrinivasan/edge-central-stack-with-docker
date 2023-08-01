// Copyright (C) 2022 IOTech Ltd

package app

import (
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms/xpert"
)

const (
	Base64Encode = "base64encode"
	Base64Decode = "base64decode"
)

// Decompress decode the binary data with base64 encoded string and decompresses from the specified algorithm (GZIP or ZLIB)
// This function is a configuration function and returns a function pointer.
func (app *Configurable) Decompress(parameters map[string]string) interfaces.AppFunction {
	algorithm, ok := parameters[Algorithm]
	if !ok {
		app.lc.Errorf("Could not find '%s' parameter for Compress", Algorithm)
		return nil
	}
	base64Decode, ok := parameters[Base64Decode]
	if !ok {
		base64Decode = "true"
	}
	base64DecodeValue, err := strconv.ParseBool(base64Decode)
	if err != nil {
		app.lc.Errorf(
			"Could not parse '%s' to boolean for '%s' parameter, %s",
			base64Decode, Base64Decode, err.Error())
		return nil
	}

	transform := xpert.Decompression{Base64Decode: base64DecodeValue}

	switch strings.ToLower(algorithm) {
	case CompressGZIP:
		return transform.DecompressWithGZIP
	case CompressZLIB:
		return transform.DecompressWithZLIB
	default:
		app.lc.Errorf(
			"Invalid decompression algorithm '%s'. Must be '%s' or '%s'",
			algorithm,
			CompressGZIP,
			CompressZLIB)
		return nil
	}
}
