package SyncMap

import (
	"errors"
	"gopher/ISyncMap"
	"sync"
)

type rwMutexMap[KEY comparable, VALUE any] struct {
	mapLock sync.RWMutex
	coreMap map[KEY]VALUE
}

func (m *rwMutexMap[KEY, VALUE]) Get(key KEY) (VALUE, error) {
	m.mapLock.RLock()
	defer m.mapLock.RUnlock()
	value, ok := m.coreMap[key]
	if ok {
		return value, nil
	} else {
		return value, errors.New("find value error")
	}
}

func (m *rwMutexMap[KEY, VALUE]) Set(key KEY, value VALUE) {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	m.coreMap[key] = value
}

func (m *rwMutexMap[KEY, VALUE]) Delete(key KEY) {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	delete(m.coreMap, key)
}

func (m *rwMutexMap[KEY, VALUE]) GetKeys() (keys []KEY) {
	m.mapLock.RLock()
	defer m.mapLock.RUnlock()
	for key, _ := range m.coreMap {
		keys = append(keys, key)
	}
	return
}

func NewRWMutexMap[KEY comparable, VALUE any]() ISyncMap.ISyncMap[KEY, VALUE] {
	return &rwMutexMap[KEY, VALUE]{
		coreMap: map[KEY]VALUE{},
	}
}
