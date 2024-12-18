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
