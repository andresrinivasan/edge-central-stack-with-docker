// Copyright (C) 2022-2023 IOTech Ltd

package xpert

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/sparkplug/protobuf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
	"time"
)

const (
	testNamespace  = "namespace"
	testGroup      = "group"
	testEdgeNodeId = "edgenodeid"
)

func TestConvertDDATAtoEvent(t *testing.T) {
	seq := uint64(1)
	ts := uint64(time.Now().UnixMilli())
	tagBytes, err := json.Marshal(existingTags)
	require.NoError(t, err, "unexpected error occurs while converting tags to []byte")
	payload := protobuf.Payload{Timestamp: &ts, Seq: &seq, Body: tagBytes}
	payloadBytes, err := proto.Marshal(&payload)
	require.NoError(t, err, "unexpected error occurs while converting sparkplug payload to []byte")
	validTopic := fmt.Sprintf("%s/%s/%s/%s/%s", testNamespace, testGroup, SparkplugBTopicLevelDDATA, testEdgeNodeId, testDeviceName)
	invalidTopicWrongType := fmt.Sprintf("%s/%s/DBIRTH/%s/%s", testNamespace, testGroup, testEdgeNodeId, testDeviceName)
	invalidTopicWrongLevel := fmt.Sprintf("%s/%s/%s/%s/%s/excrescentTopicLevel", testNamespace, testGroup, SparkplugBTopicLevelDDATA, testEdgeNodeId, testDeviceName)

	tests := []struct {
		Name             string
		inputData        interface{}
		receivedTopic    string
		ErrorExpectation bool
	}{
		{"No inputData", nil, validTopic, true},
		{"inputData in wrong format", dtos.Event{}, validTopic, true},
		{"invalid topic - wrong type", payloadBytes, invalidTopicWrongType, true},
		{"invalid topic - wrong level", payloadBytes, invalidTopicWrongLevel, true},
		{"Correct inputData", payloadBytes, validTopic, false},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx.AddValue(interfaces.RECEIVEDTOPIC, test.receivedTopic)
			continuePipeline, result := NewSparkplugConverter().ConvertDDATAtoEvent(ctx, test.inputData)
			if test.ErrorExpectation {
				assert.Equal(t, continuePipeline, false)
			} else {
				assert.Equal(t, continuePipeline, true)
				event, typeAccepted := result.(dtos.Event)
				assert.True(t, typeAccepted, "result returned is not an Event")
				assert.NotZero(t, event.Tags, "empty event tags")
				for tag, tagValue := range existingTags {
					val, ok := event.Tags[tag]
					assert.True(t, ok)
					assert.Equal(t, tagValue, val)
				}
			}
		})
	}
}
