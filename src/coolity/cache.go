package coolify

import (
	"sync"
	"time"
)

type cacheItem struct {
	data      any
	expiresAt time.Time
}

type cache struct {
	items sync.Map
	ttl   time.Duration
}

func newCache(ttl time.Duration) *cache {
	return &cache{
		ttl: ttl,
	}
}

func (c *cache) Get(key string) (any, bool) {
	item, ok := c.items.Load(key)
	if !ok {
		return nil, false
	}

	cacheItem := item.(cacheItem)
	if time.Now().After(cacheItem.expiresAt) {
		c.items.Delete(key)
		return nil, false
	}

	return cacheItem.data, true
}

func (c *cache) Set(key string, value any) {
	c.items.Store(key, cacheItem{
		data:      value,
		expiresAt: time.Now().Add(c.ttl),
	})
}

func (c *cache) Delete(key string) {
	c.items.Delete(key)
}

func (c *cache) Clear() {
	c.items.Range(func(key, value any) bool {
		c.items.Delete(key)
		return true
	})
}
