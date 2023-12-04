package cache

import (
	"sync"
	"time"
)

type LRUCache[K comparable, V any] struct {
	mx *sync.RWMutex
	// time that an item can live in the cache, in the second.
	maximumLifeTime int64
	data            map[K]value[V]
}

type value[V any] struct {
	lastUsedTime int64
	value        V
}

func NewLRUCache[K comparable, V any](
	maximumLifeTime int64,
) *LRUCache[K, V] {
	result := &LRUCache[K, V]{
		mx:              new(sync.RWMutex),
		maximumLifeTime: maximumLifeTime,
		data:            make(map[K]value[V]),
	}

	result.start()

	return result
}

func (c *LRUCache[K, V]) start() {
	go func() {
		for {
			c.mx.Lock()
			for k, v := range c.data {
				if v.lastUsedTime+c.maximumLifeTime < time.Now().Unix() {
					delete(c.data, k)
				}
			}
			c.mx.Unlock()

			time.Sleep(time.Duration(c.maximumLifeTime/2) * time.Second)
		}
	}()
}

func (c *LRUCache[K, V]) Set(key K, val V) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.data[key] = value[V]{
		lastUsedTime: time.Now().Unix(),
		value:        val,
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	v, exists := c.data[key]

	return v.value, exists
}

func (c *LRUCache[K, V]) Delete(key K) {
	c.mx.Lock()
	defer c.mx.Unlock()

	delete(c.data, key)
}

func (c *LRUCache[K, V]) Len() int {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return len(c.data)
}

func (c *LRUCache[K, V]) Clear() {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.data = make(map[K]value[V])
}
