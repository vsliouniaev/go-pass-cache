package main

import (
	"sync"
	"time"
)

type Cache struct {
	data map[string]CacheObject
	mx   sync.RWMutex
}

type CacheObject struct {
	data    string
	created time.Time
}

func (c *Cache) TryGetAndRemoveWithinTimeFrame(key string, duration time.Duration) (string, bool) {
	if _, ok := c.data[key]; !ok {
		return "", false
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	if val, ok := c.data[key]; ok {
		delete(c.data, key)
		if time.Since(val.created) < duration {
			return val.data, true
		}
	}

	return "", false
}

func (c *Cache) AddOrSilentlyFail(key string, val string) {
	if _, ok := c.data[key]; ok {
		return
	}

	c.mx.Lock()
	defer c.mx.Unlock()
	if _, ok := c.data[key]; !ok {
		c.data[key] = CacheObject{val, time.Now()}
	}
}
