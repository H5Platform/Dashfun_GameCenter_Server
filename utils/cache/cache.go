package cache

import (
	"sync"
	"time"
)

// item 是泛型缓存中的条目，包含值和过期时间
type item[T any] struct {
	value     T
	expiresAt time.Time
}

// GenericCache 是带 TTL 的泛型缓存结构
type GenericCache[T any] struct {
	mu    sync.RWMutex
	items map[string]item[T]
	ttl   time.Duration
}

// NewCache 创建一个新的泛型缓存，带默认 TTL
func NewCache[T any](ttl time.Duration) *GenericCache[T] {
	c := &GenericCache[T]{
		items: make(map[string]item[T]),
		ttl:   ttl,
	}
	go c.startEvictionLoop()
	return c
}

// Set 写入缓存项
func (c *GenericCache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item[T]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Get 读取缓存项（如果存在且未过期）
func (c *GenericCache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.items[key]
	if !ok || time.Now().After(it.expiresAt) {
		var zero T
		return zero, false
	}
	return it.value, true
}

// Delete 删除某个缓存项
func (c *GenericCache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len 返回当前缓存的大小
func (c *GenericCache[T]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Flush 清空所有缓存
func (c *GenericCache[T]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]item[T])
}

// startEvictionLoop 启动后台协程定期清理过期项
func (c *GenericCache[T]) startEvictionLoop() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for range ticker.C {
		c.evictExpired()
	}
}

// evictExpired 执行过期清理
func (c *GenericCache[T]) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, it := range c.items {
		if now.After(it.expiresAt) {
			delete(c.items, k)
		}
	}
}
