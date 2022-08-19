package cache

import (
	"sync"
	"time"
)

// Cache is a write and read-at-most-once store for data.
type Cache interface {
	// TryGet will retrieve previously stored value if it has not yet expired and if it has not yet been accessed.
	TryGet(key string) (string, bool)
	// Store will write data into the cache.
	Store(key string, val string)
}

// New creates a new Cache where items older than duration will not be retrieved.
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

var _ Cache = &cache{}

// cache implements the Cache interface. To simplify expiration of data two maps are used and rotated every duration.
// A key may end up being stored for 2x the duration but will not be returned because its timestamp is validated on
// retrieval.
type cache struct {
	data1    map[string]*cached
	data2    map[string]*cached
	duration time.Duration
	mx       sync.RWMutex
}

// cached represents data, with its creation timestamp ensuring it will be correctly expired.
type cached struct {
	data    string
	created time.Time
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

func (c *cache) Store(key string, val string) {
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

func (c *cache) keyExists(key string) bool {
	c.mx.RLock()
	defer c.mx.RUnlock()
	_, ok := c.getOneOrTwo(key)
	return ok
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
