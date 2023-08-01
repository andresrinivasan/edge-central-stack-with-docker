// Copyright (C) 2022 IOTech Ltd

package xpert

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetContextVariable(t *testing.T) {
	testVariableName := "var"
	testValueJsonPath := "tags.var-name"
	testValidJsonString := "{\n    \"apiVersion\": \"v2\",\n    \"id\": \"040bd523-ec33-440d-9d72-e5813a465f37\",\n    \"origin\": 1602168089665565200,\n    \"deviceName\": \"device-001\",\n    \"profileName\": \"profile-001\",\n    \"sourceName\": \"source-1\",\n    \"tags\":\n    {\n        \"var-name\": \"value\",\n        \"interface-name\": \"interface-001\"\n    }\n}"
	testValidNotFoundJsonString := "{\n    \"apiVersion\": \"v2\",\n    \"id\": \"040bd523-ec33-440d-9d72-e5813a465f37\",\n    \"origin\": 1602168089665565200,\n    \"deviceName\": \"device-001\",\n    \"profileName\": \"profile-001\",\n    \"sourceName\": \"source-1\",\n    \"tags\":\n    {\n        \"var\": \"value\",\n        \"interface-name\": \"interface-001\"\n    }\n}"
	expectedJsonPathValue := "value"
	testInvalidJsonString := "[{abc]"
	expectedJsonPathNumberValue := "1602168089665565200"

	tests := []struct {
		Name              string
		variableName      string
		valueJsonPath     string
		continueOnError   bool
		data              interface{}
		expectedJsonValue string
		ErrorExpectation  bool
	}{
		{"Normal case with valid params - String", testVariableName, testValueJsonPath, false, testValidJsonString, expectedJsonPathValue, false},
		{"Normal case with valid params - Number", testVariableName, "origin", false, testValidJsonString, expectedJsonPathNumberValue, false},
		{"Normal case with valid params but not found", testVariableName, testValueJsonPath, false, testValidNotFoundJsonString, "", false},
		{"Error case with invalid json string", testVariableName, testValueJsonPath, false, testInvalidJsonString, "", true},
		{"Error case with invalid json string and continueOnError is true", testVariableName, testValueJsonPath, true, testInvalidJsonString, expectedJsonPathValue, false},
		{"Error case with nil data and continueOnError is true", testVariableName, testValueJsonPath, true, nil, expectedJsonPathValue, false},
		{"Error case with nil data and continueOnError is false", testVariableName, testValueJsonPath, false, nil, expectedJsonPathValue, true},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			setter := NewContextVariableSetter(test.variableName, test.valueJsonPath, test.continueOnError)
			continuePipeline, result := setter.SetContextVariable(ctx, test.data)
			if test.ErrorExpectation {
				assert.Equal(t, continuePipeline, false)
			} else {
				assert.Equal(t, continuePipeline, true)
				assert.Equal(t, test.data, result)
				if !test.continueOnError {
					ctxVal, found := ctx.GetValue(test.variableName)
					assert.True(t, found, "Should find Context Variable")
					assert.Equal(t, test.expectedJsonValue, ctxVal, "Context Variable value should be identical")
				}
			}
		})
	}
}
