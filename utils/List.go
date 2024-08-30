package utils

import (
	"reflect"
	"sync"
)

type List[T any] interface {
	Add(item T)
	Remove(item T) bool
	RemoveAt(index int) bool
	Get(index int) T
	IndexOf(item T) int
	Size() int
	Items() []T
}

type baseList[T any] struct {
	items []T
}

type syncList[T any] struct {
	list List[T]
	sync.RWMutex
}

func (b *baseList[T]) Add(item T) {
	b.items = append(b.items, item)
}

func (b *baseList[T]) Remove(item T) bool {
	var idx = b.IndexOf(item)
	if idx == -1 {
		return false
	}
	b.items = append(b.items[:idx], b.items[idx+1:]...)
	return true
}

func (b *baseList[T]) RemoveAt(index int) bool {
	if index < 0 || index >= len(b.items) {
		return false
	}
	b.items = append(b.items[:index], b.items[index+1:]...)
	return true
}

func (b *baseList[T]) Get(index int) T {
	return b.items[index]
}

func (b *baseList[T]) IndexOf(item T) int {
	for i, v := range b.items {
		if reflect.DeepEqual(v, item) {
			return i
		}
	}
	return -1
}

func (b *baseList[T]) Size() int {
	return len(b.items)
}

func (b *baseList[T]) Items() []T {
	return b.items
}

func (s *syncList[T]) Add(item T) {
	s.Lock()
	defer s.Unlock()
	s.list.Add(item)
}

func (s *syncList[T]) Remove(item T) bool {
	s.Lock()
	defer s.Unlock()
	return s.list.Remove(item)
}

func (s *syncList[T]) Get(index int) T {
	s.RLock()
	defer s.RUnlock()
	return s.list.Get(index)
}

func (s *syncList[T]) IndexOf(item T) int {
	s.RLock()
	defer s.RUnlock()
	return s.list.IndexOf(item)
}

func (s *syncList[T]) Size() int {
	s.Lock()
	defer s.Unlock()
	return s.list.Size()
}

func (s *syncList[T]) Items() []T {
	s.RLock()
	defer s.RUnlock()
	return s.list.Items()
}

func (s *syncList[T]) RemoveAt(index int) bool {
	s.Lock()
	defer s.Unlock()
	return s.list.RemoveAt(index)
}

func NewList[T any]() List[T] {
	ret := &baseList[T]{
		items: make([]T, 0),
	}
	return ret
}

func NewSynchronizedList[T any]() List[T] {
	ret := &syncList[T]{
		list: NewList[T](),
	}
	return ret
}
