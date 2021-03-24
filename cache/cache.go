package cache

import (
	"sync"
	"time"
)

type Cache interface {
	TryGet(key string) (string, bool)
	AddKey(key string, val string)
}

type cache struct {
	data1    map[string]*cached
	data2    map[string]*cached
	duration time.Duration
	mx       sync.RWMutex
}

func New(duration time.Duration) Cache {
	c := &cache{
		data1:    make(map[string]*cached),
		duration: duration,
	}

	ticker := time.NewTicker(duration)
	go func() {
		for {
			<-ticker.C
			c.rotate()
		}
	}()

	return c
}

type cached struct {
	data    string
	created time.Time
}

func (c *cache) keyExists(key string) bool {
	c.mx.RLock()
	defer c.mx.RUnlock()
	_, ok := c.getOneOrTwo(key)
	return ok
}

func (c *cache) TryGet(key string) (string, bool) {
	if !c.keyExists(key) {
		return "", false
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	if val, ok := c.getOneOrTwo(key); ok {
		c.delete(key)
		if time.Since(val.created) < c.duration {
			return val.data, true
		}
	}
	return "", false
}

// If there is no key, add it. If the key exists already, delete it without replacing.
func (c *cache) AddKey(key string, val string) {
	obj := &cached{val, time.Now()}
	c.mx.Lock()
	defer c.mx.Unlock()
	if _, ok := c.getOneOrTwo(key); !ok {
		c.data1[key] = obj
	} else {
		// If there's a clash for some reason just be safe and delete it
		c.delete(key)
	}
}

func (c *cache) getOneOrTwo(key string) (*cached, bool) {
	val, ok := c.data1[key]
	if !ok {
		val, ok = c.data2[key]
		if !ok {
			return nil, false
		}
	}
	return val, true
}

func (c *cache) delete(key string) {
	delete(c.data1, key)
	delete(c.data2, key)
}

func (c *cache) rotate() {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.data2 = c.data1
	c.data1 = make(map[string]*cached)
}
