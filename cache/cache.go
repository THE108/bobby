package cache

import (
	"sync"
	"time"
)

type cacheItem struct {
	createdAt time.Time
	ttl       time.Duration
	value     string
}

type Cache struct {
	lock sync.RWMutex
	data map[string]cacheItem
}

func NewCache(capacity int) *Cache {
	return &Cache{
		data: make(map[string]cacheItem, capacity),
	}
}

func (this *Cache) Get(key string) (string, bool) {
	this.lock.RLock()
	item, found := this.data[key]
	this.lock.RUnlock()

	if !found {
		return "", false
	}

	if item.createdAt.Add(item.ttl).Before(time.Now()) {
		return "", false
	}

	return item.value, true
}

func (this *Cache) Set(key string, value string, ttl time.Duration) {
	now := time.Now()
	this.lock.Lock()
	this.data[key] = cacheItem{
		createdAt: now,
		ttl:       ttl,
		value:     value,
	}
	this.lock.Unlock()
}
