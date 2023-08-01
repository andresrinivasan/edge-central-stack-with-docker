// Copyright (C) 2023 IOTech Ltd

package xpert

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	defaultMax       = 10
	defaultInitValue = 0
)

func TestCounterSet(t *testing.T) {
	tests := []struct {
		name        string
		maximum     uint
		initValue   uint
		setValue    uint
		expectError bool
	}{
		{"Normal set below maximum", defaultMax, defaultInitValue, 5, false},
		{"Set above maximum", defaultMax, defaultInitValue, defaultMax + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewCounter(tt.maximum, tt.initValue)
			err := counter.Set(tt.setValue)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.setValue, counter.Value())
			}
		})
	}
}

func TestCounterAdd(t *testing.T) {
	tests := []struct {
		name        string
		maximum     uint
		initValue   uint
		delta       uint
		expectError bool
	}{
		{"Normal add below maximum", defaultMax, defaultInitValue, 5, false},
		{"Add above maximum", defaultMax, defaultInitValue, defaultMax + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewCounter(tt.maximum, tt.initValue)
			init := counter.Value()
			assert.Equal(t, tt.initValue, init)
			err := counter.Add(tt.delta)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, init+tt.delta, counter.Value())
			}
		})
	}
}

func TestCounterSubtract(t *testing.T) {
	tests := []struct {
		name      string
		maximum   uint
		initValue uint
		delta     uint
	}{
		{"Normal subtract above zero", defaultMax, defaultInitValue + 2, 1},
		{"Subtract below zero", defaultMax, defaultInitValue, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewCounter(tt.maximum, tt.initValue)
			init := counter.Value()
			assert.Equal(t, tt.initValue, init)
			expected := int(init) - int(tt.delta)
			if expected < 0 {
				expected = 0
			}
			counter.Subtract(tt.delta)
			assert.Equal(t, uint(expected), counter.Value())
		})
	}
}
