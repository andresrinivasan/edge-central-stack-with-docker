// Copyright (C) 2023 IOTech Ltd

package xpert

import (
	"fmt"
	"sync"
)

const DefaultCounterMaximum = uint(100)

// AtomicCounter is counter that can guarantee all its operation, Value(), Set(), Add(), Subtract(), in atomic execution.
type AtomicCounter struct {
	mutex   sync.Mutex
	maximum uint
	count   uint
}

// NewCounter will return a counter instance with specified maximum and initial counter value.
func NewCounter(maximum uint, initCount uint) *AtomicCounter {
	return &AtomicCounter{
		maximum: maximum,
		count:   initCount,
	}
}

// Value return the current value of counter
func (i *AtomicCounter) Value() uint {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	v := i.count
	return v
}

// Set reset the value of counter.  Return error when attempting to set counter to value above maximum
func (i *AtomicCounter) Set(v uint) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if v > i.maximum {
		return fmt.Errorf("fail to set counter to %d as this operation will exceed counter maximum %d", v, i.maximum)
	}
	i.count = v
	return nil
}

// Add adds the counter with delta.  Return error when the attempted increment will result in the
// value of counter exceeding the maximum; return nil when the increment is successful.
func (i *AtomicCounter) Add(delta uint) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	newCount := i.count + delta
	if newCount > i.maximum {
		return fmt.Errorf("fail to add counter as this increment operation will exceed counter maximum %d", i.maximum)
	}
	i.count = newCount
	return nil
}

// Subtract minuses the value of counter with delta.  Note that if the decrement operation will result in a negative
// value of counter, this function will automatically set the counter to zero.
func (i *AtomicCounter) Subtract(delta uint) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	newCount := int(i.count) - int(delta)
	if newCount < 0 {
		newCount = 0
	}
	i.count = uint(newCount)
}
