# gopher

## 使用读写锁和通道分别实现线程安全的map

### SyncMap接口

```go
type ISyncMap[KEY comparable, VALUE any] interface {
	Get(KEY) (VALUE, error)//获得一个键对应的值
	Set(KEY, VALUE)//设置一个键值对
	Delete(KEY)//删除一个键值对
	GetKeys() (keys []KEY)//获得所有的键列表
}
```

### 使用读写锁

```go
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

```

### 使用通道

```go
package SyncMap

import (
	"errors"
	"gopher/ISyncMap"
)

type order[KEY comparable, VALUE any] struct {
	orderType  int
	key        KEY
	value      VALUE
	resultChan chan result[KEY, VALUE]
}

type result[KEY comparable, VALUE any] struct {
	value VALUE
	keys  []KEY
	err   error
}

type channelMap[KEY comparable, VALUE any] struct {
	coreMap      map[KEY]VALUE
	orderChannel chan order[KEY, VALUE]
}

func (c *channelMap[KEY, VALUE]) receiveOrder() {
	for {
		o := <-c.orderChannel
		if o.orderType == 1 {
			//当指令类型为1时，为添加一个键值对
			c.coreMap[o.key] = o.value
		} else if o.orderType == 2 {
			//当指令类型为2时，获得一个键对应的值
			value, ok := c.coreMap[o.key]
			r := result[KEY, VALUE]{}
			if ok {
				r.value = value
			} else {
				r.err = errors.New("find value error")
			}
			o.resultChan <- r
		} else if o.orderType == 3 {
			//当指令类型为3时，删除一个键值对
			delete(c.coreMap, o.key)
		} else if o.orderType == 4 {
			//当指令类型为3时，获得一个键的列表
			r := result[KEY, VALUE]{keys: make([]KEY, 0)}
			for key, _ := range c.coreMap {
				r.keys = append(r.keys, key)
			}
			o.resultChan <- r
		}
	}
}

func NewChannelMap[KEY comparable, VALUE any]() ISyncMap.ISyncMap[KEY, VALUE] {
	c := channelMap[KEY, VALUE]{
		coreMap:      make(map[KEY]VALUE),
		orderChannel: make(chan order[KEY, VALUE]),
	}
	go c.receiveOrder()
	return &c
}

func (c *channelMap[KEY, VALUE]) Set(key KEY, value VALUE) {
	o := order[KEY, VALUE]{
		orderType: 1,
		key:       key,
		value:     value,
	}
	c.orderChannel <- o
}

func (c *channelMap[KEY, VALUE]) Get(key KEY) (VALUE, error) {
	o := order[KEY, VALUE]{
		orderType:  2,
		key:        key,
		resultChan: make(chan result[KEY, VALUE]),
	}
	c.orderChannel <- o
	r := <-o.resultChan
	return r.value, r.err
}

func (c *channelMap[KEY, VALUE]) Delete(key KEY) {
	o := order[KEY, VALUE]{
		orderType: 3,
		key:       key,
	}
	c.orderChannel <- o
}

func (c *channelMap[KEY, VALUE]) GetKeys() []KEY {
	o := order[KEY, VALUE]{
		orderType:  4,
		resultChan: make(chan result[KEY, VALUE]),
	}
	c.orderChannel <- o
	r := <-o.resultChan
	return r.keys
}

```

