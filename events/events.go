package events

import (
	"reflect"
	"sync"
)

type event[T any] interface {
	On(listener func(T))
	Remove(listener func(T))
	Emit(T)
}

type EventListenerList[T any] struct {
	sync.RWMutex
	list []EventListenerChan[T]
}

type EventListenerChan[T any] struct {
	l  func(T)
	ch chan T
}

func (l *EventListenerList[T]) Add(listener func(T)) {
	l.Lock()
	defer l.Unlock()

	ec := EventListenerChan[T]{
		l:  listener,
		ch: make(chan T, 1),
	}

	l.list = append(l.list, ec)

	go func() {
		for msg := range ec.ch {
			//zap.S().Infow("fire event", "msg", msg)
			ec.l(msg)
		}
	}()
}

func (l *EventListenerList[T]) Remove(listener func(T)) {
	l.Lock()
	defer l.Unlock()
	for i, e := range l.list {
		ptr1 := reflect.ValueOf(e.l)
		ptr2 := reflect.ValueOf(listener)
		if ptr1.Pointer() == ptr2.Pointer() {
			l.list = append(l.list[:i], l.list[i+1:]...)
			close(e.ch)
		}
	}
}

type EventBase[T any] struct {
	Listeners EventListenerList[T]
}

func (e *EventBase[T]) On(listener func(T)) {
	e.Listeners.Add(listener)
}

func (e *EventBase[T]) Remove(listener func(T)) {
	e.Listeners.Remove(listener)
}

func (e *EventBase[T]) Emit(t T) {
	for _, l := range e.Listeners.list {
		//l(t)
		select {
		case l.ch <- t:
		default:
			return
		}
	}
}

func NewEvent[T any]() *EventBase[T] {
	e := &EventBase[T]{
		Listeners: EventListenerList[T]{},
	}
	return e
}
