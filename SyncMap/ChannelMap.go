package SyncMap

import (
	"errors"
	"gopher/ISyncMap"
	"sync"
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
	coreMap        map[KEY]VALUE
	orderChannel   chan order[KEY, VALUE]
	orderPool      sync.Pool
	resultPool     sync.Pool
	resultChanPool sync.Pool
}

func (c *channelMap[KEY, VALUE]) receiveOrder() {
	defer close(c.orderChannel) //将指令通道关闭
	for {
		o := <-c.orderChannel
		if o.orderType == 1 {
			//当指令类型为1时，为添加一个键值对
			c.coreMap[o.key] = o.value
		} else if o.orderType == 2 {
			//当指令类型为2时，获得一个键对应的值
			value, ok := c.coreMap[o.key]
			r := c.resultPool.Get().(*result[KEY, VALUE])
			if ok {
				r.value = value
			} else {
				r.err = errors.New("find value error")
			}
			o.resultChan <- *r
			c.resultPool.Put(r)
		} else if o.orderType == 3 {
			//当指令类型为3时，删除一个键值对
			delete(c.coreMap, o.key)
		} else if o.orderType == 4 {
			//当指令类型为3时，获得一个键的列表
			r := c.resultPool.Get().(*result[KEY, VALUE])
			for key, _ := range c.coreMap {
				r.keys = append(r.keys, key)
			}
			o.resultChan <- *r
			c.resultPool.Put(r)
		}
	}
}

func NewChannelMap[KEY comparable, VALUE any]() ISyncMap.ISyncMap[KEY, VALUE] {
	c := channelMap[KEY, VALUE]{
		coreMap:      make(map[KEY]VALUE),
		orderChannel: make(chan order[KEY, VALUE]),
		orderPool: sync.Pool{New: func() interface{} {
			return new(order[KEY, VALUE])
		},
		},
		resultPool: sync.Pool{New: func() interface{} {
			return new(result[KEY, VALUE])
		},
		},
		resultChanPool: sync.Pool{New: func() interface{} {
			return make(chan result[KEY, VALUE])
		},
		},
	}
	go c.receiveOrder()
	return &c
}

func (c *channelMap[KEY, VALUE]) Set(key KEY, value VALUE) {
	o := c.orderPool.Get().(*order[KEY, VALUE])
	o.orderType = 1
	o.key = key
	o.value = value
	c.orderChannel <- *o
	c.orderPool.Put(o)
}

func (c *channelMap[KEY, VALUE]) Get(key KEY) (VALUE, error) {
	o := c.orderPool.Get().(*order[KEY, VALUE])
	o.orderType = 2
	o.key = key
	o.resultChan = c.resultChanPool.Get().(chan result[KEY, VALUE])
	c.orderChannel <- *o
	r := <-o.resultChan
	c.resultChanPool.Put(o.resultChan)
	return r.value, r.err
}

func (c *channelMap[KEY, VALUE]) Delete(key KEY) {
	o := c.orderPool.Get().(*order[KEY, VALUE])
	o.orderType = 3
	o.key = key
	c.orderChannel <- *o
	c.orderPool.Put(o)
}

func (c *channelMap[KEY, VALUE]) GetKeys() []KEY {
	o := c.orderPool.Get().(*order[KEY, VALUE])
	o.orderType = 4
	o.resultChan = c.resultChanPool.Get().(chan result[KEY, VALUE])
	c.orderChannel <- *o
	r := <-o.resultChan
	c.orderPool.Put(o)
	c.resultChanPool.Put(o.resultChan)
	return r.keys
}
