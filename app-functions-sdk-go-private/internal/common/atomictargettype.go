// Copyright (C) 2021 IOTech Ltd

package common

import "sync"

type AtomicTargetType struct {
	targetType interface{}
	mutex      sync.RWMutex
}

func NewAtomicTargetType(targetType interface{}) *AtomicTargetType {
	result := AtomicTargetType{}
	result.Set(targetType)
	return &result
}

func (att *AtomicTargetType) IsNil() bool {
	att.mutex.RLock()
	defer att.mutex.RUnlock()
	return att.targetType == nil
}

func (att *AtomicTargetType) Set(targetType interface{}) {
	att.mutex.Lock()
	defer att.mutex.Unlock()
	att.targetType = targetType
}

func (att *AtomicTargetType) Type() interface{} {
	att.mutex.RLock()
	defer att.mutex.RUnlock()
	return att.targetType
}
