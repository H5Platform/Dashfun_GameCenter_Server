package utils

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

// lockEntry 包含具体的 RWMutex、引用计数和清理定时器
type lockEntry struct {
	lock     *sync.RWMutex
	refCount int
	timer    *time.Timer
}

// KeyRWMutexManager 支持任意 comparable key 的 RWMutex 管理器，具有引用计数和过期清理功能
type KeyRWMutexManager[K comparable] struct {
	name        string
	mu          sync.Mutex
	locks       map[K]*lockEntry
	expireAfter time.Duration
}

// LockGuard 封装获取的锁和解锁 + 释放引用逻辑
type LockGuard[K comparable] struct {
	key    K
	mgr    *KeyRWMutexManager[K]
	unlock func()
}

func (g *LockGuard[K]) Unlock() {
	g.unlock()
	g.mgr.Release(g.key)
}

// NewKeyRWMutexManager 创建一个带自动过期清理的锁管理器
func NewKeyRWMutexManager[K comparable](name string, expireAfter time.Duration) *KeyRWMutexManager[K] {
	return &KeyRWMutexManager[K]{
		name:        name,
		locks:       make(map[K]*lockEntry),
		expireAfter: expireAfter,
	}
}

// Get 获取某个 key 对应的锁（并增加引用）
func (mgr *KeyRWMutexManager[K]) Get(key K) *sync.RWMutex {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	entry, ok := mgr.locks[key]
	if !ok {
		entry = &lockEntry{
			lock:     &sync.RWMutex{},
			refCount: 1,
		}
		mgr.locks[key] = entry
	} else {
		entry.refCount++
		if entry.timer != nil {
			entry.timer.Stop()
			entry.timer = nil
		}
	}
	return entry.lock
}

// Lock 获取写锁并返回封装的 LockGuard
func (mgr *KeyRWMutexManager[K]) Lock(key K) *LockGuard[K] {
	lock := mgr.Get(key)
	lock.Lock()
	return &LockGuard[K]{
		key:    key,
		mgr:    mgr,
		unlock: lock.Unlock,
	}
}

// RLock 获取读锁并返回封装的 LockGuard
func (mgr *KeyRWMutexManager[K]) RLock(key K) *LockGuard[K] {
	lock := mgr.Get(key)
	lock.RLock()
	return &LockGuard[K]{
		key:    key,
		mgr:    mgr,
		unlock: lock.RUnlock,
	}
}

// Release 减少锁引用，并在引用为 0 时延迟清理
func (mgr *KeyRWMutexManager[K]) Release(key K) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	entry, ok := mgr.locks[key]
	if !ok {
		return
	}

	entry.refCount--
	if entry.refCount > 0 {
		return
	}

	zap.S().Debugf("KeyRWMutexManager-[%s] key %v released, scheduling cleanup. active keys: %d", mgr.name, key, len(mgr.locks)-1)

	entry.timer = time.AfterFunc(mgr.expireAfter, func() {
		mgr.mu.Lock()
		defer mgr.mu.Unlock()

		if e, still := mgr.locks[key]; still && e == entry && e.refCount == 0 {
			delete(mgr.locks, key)
			zap.S().Debugf("KeyRWMutexManager-[%s] key %v expired and removed. active keys: %d", mgr.name, key, len(mgr.locks))
		}
	})
}
