package events

import (
	"reflect"
	"sync"
)

type event[T any] interface {
	On(listener func(T))
	Emit(T)
}

type EventListenerList[T any] struct {
	sync.RWMutex
	list []func(T)
}

func (l *EventListenerList[T]) Add(listener func(T)) {
	l.Lock()
	defer l.Unlock()
	l.list = append(l.list, listener)
}

func (l *EventListenerList[T]) Remove(listener func(T)) {
	l.Lock()
	defer l.Unlock()
	for i, e := range l.list {
		ptr1 := reflect.ValueOf(e)
		ptr2 := reflect.ValueOf(listener)
		if ptr1.Pointer() == ptr2.Pointer() {
			l.list = append(l.list[:i], l.list[i+1:]...)
		}
	}
}

type EventBase[T any] struct {
	Listeners EventListenerList[T]
}

func (e *EventBase[T]) On(listener func(T)) {
	e.Listeners.Add(listener)
}

func (e *EventBase[T]) Emit(t T) {
	for _, l := range e.Listeners.list {
		l(t)
	}
}

func NewEvent[T any]() *EventBase[T] {
	e := &EventBase[T]{
		Listeners: EventListenerList[T]{},
	}
	return e
}
