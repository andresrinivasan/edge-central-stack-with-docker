// Copyright (C) 2022 IOTech Ltd

package xpert

import (
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	clearString = "This is the test string used for testing"
)

func TestCompressWithGZIP(t *testing.T) {
	comp := transforms.NewCompression()
	continuePipeline, compressed := comp.CompressWithGZIP(ctx, []byte(clearString))
	assert.True(t, continuePipeline)

	decomp := NewDecompression()
	continuePipeline, decompressed := decomp.DecompressWithGZIP(ctx, compressed)
	assert.True(t, continuePipeline)
	res, err := util.CoerceType(decompressed)
	assert.NoError(t, err)
	require.Equal(t, clearString, string(res))
}

func TestCompressWithZLIB(t *testing.T) {
	comp := transforms.NewCompression()
	continuePipeline, compressed := comp.CompressWithZLIB(ctx, []byte(clearString))
	assert.True(t, continuePipeline)

	decomp := NewDecompression()
	continuePipeline, decompressed := decomp.DecompressWithZLIB(ctx, compressed)
	assert.True(t, continuePipeline)
	res, err := util.CoerceType(decompressed)
	assert.NoError(t, err)
	require.Equal(t, clearString, string(res))
}

func BenchmarkGZIPDecompress(b *testing.B) {
	comp := transforms.NewCompression()
	continuePipeline, compressed := comp.CompressWithGZIP(ctx, []byte(clearString))
	assert.True(b, continuePipeline)

	decomp := NewDecompression()

	var decompressed interface{}
	for i := 0; i < b.N; i++ {
		_, decompressed = decomp.DecompressWithGZIP(ctx, compressed)
	}
	b.SetBytes(int64(len(decompressed.([]byte))))
}

func BenchmarkZLIBDecompress(b *testing.B) {
	comp := transforms.NewCompression()
	continuePipeline, compressed := comp.CompressWithZLIB(ctx, []byte(clearString))
	assert.True(b, continuePipeline)

	decomp := NewDecompression()

	var decompressed interface{}
	for i := 0; i < b.N; i++ {
		_, decompressed = decomp.DecompressWithZLIB(ctx, compressed)
	}
	b.SetBytes(int64(len(decompressed.([]byte))))
}
