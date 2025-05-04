package utils

import "sync"

type Map[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Exists(key K) bool
	Len() int
	Range(f func(key K, value V) bool)
	Size() int
}

func NewSynchronizedMap[K comparable, V any]() Map[K, V] {
	return &SynchronizedMap[K, V]{m: make(map[K]V)}
}

func NewUnsafeMap[K comparable, V any]() Map[K, V] {
	return &UnsafeMap[K, V]{m: make(map[K]V)}
}

type SynchronizedMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func (s *SynchronizedMap[K, V]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

func (s *SynchronizedMap[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.m[key]
	return val, ok
}

func (s *SynchronizedMap[K, V]) Set(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *SynchronizedMap[K, V]) Delete(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}

func (s *SynchronizedMap[K, V]) Exists(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[key]
	return ok
}

func (s *SynchronizedMap[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

func (s *SynchronizedMap[K, V]) Range(f func(key K, value V) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.m {
		if !f(k, v) {
			break
		}
	}
}

type UnsafeMap[K comparable, V any] struct {
	m map[K]V
}

func (s *UnsafeMap[K, V]) Get(key K) (V, bool) {
	val, ok := s.m[key]
	return val, ok
}

func (s *UnsafeMap[K, V]) Set(key K, value V) {
	s.m[key] = value
}

func (s *UnsafeMap[K, V]) Delete(key K) {
	delete(s.m, key)
}

func (s *UnsafeMap[K, V]) Exists(key K) bool {
	_, ok := s.m[key]
	return ok
}

func (s *UnsafeMap[K, V]) Len() int {
	return len(s.m)
}

func (s *UnsafeMap[K, V]) Range(f func(key K, value V) bool) {
	for k, v := range s.m {
		if !f(k, v) {
			break
		}
	}
}
func (s *UnsafeMap[K, V]) Size() int {
	return len(s.m)
}
