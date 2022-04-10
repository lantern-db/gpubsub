package gpubsub

import (
	"sync"
	"time"
)

type Topic[T any] struct {
	mu            sync.RWMutex
	name          string
	subscriptions map[string]*Subscription[T]
	concurrency   int64
	interval      time.Duration
	ttl           time.Duration
}

func NewTopic[T any](name string, concurrency int64, interval time.Duration, ttl time.Duration) *Topic[T] {
	return &Topic[T]{
		name:          name,
		subscriptions: make(map[string]*Subscription[T]),
		concurrency:   concurrency,
		interval:      interval,
		ttl:           ttl,
	}
}

func (t *Topic[T]) Name() string {
	return t.name
}

func (t *Topic[T]) Subscriptions() map[string]*Subscription[T] {
	return t.subscriptions
}

func (t *Topic[T]) Publish(body T) {
	for _, s := range t.subscriptions {
		message := s.NewMessage(body)
		s.Publish(message)
	}
}

func (t *Topic[T]) NewSubscription(name string) *Subscription[T] {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.subscriptions[name]; !ok {
		t.subscriptions[name] = &Subscription[T]{
			name:        name,
			topic:       t,
			ch:          make(chan string, 65536),
			messages:    make(map[string]*Message[T]),
			concurrency: t.concurrency,
			interval:    t.interval,
			ttl:         t.ttl,
		}
	}
	return t.subscriptions[name]
}

func (t *Topic[T]) Register(subscription *Subscription[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.subscriptions[subscription.name]; !ok {
		t.subscriptions[subscription.name] = subscription
	}
}

func (t *Topic[T]) Unregister(subscription *Subscription[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.subscriptions[subscription.name]; ok {
		delete(t.subscriptions, subscription.name)
	}
}
