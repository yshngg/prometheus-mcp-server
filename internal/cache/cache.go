package cache

import (
	"sync"
	"time"
)

type entry struct {
	value   any
	expires time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]*entry
	ttl   time.Duration
}

func New(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]*entry),
		ttl:   ttl,
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	c.items[key] = &entry{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
