package gpubsub

import (
	"context"
	"golang.org/x/sync/semaphore"
	"log"
	"sync"
)

type Topic[T any] struct {
	mu            sync.RWMutex
	name          string
	subscriptions map[string]*Subscription[T]
}

func NewTopic[T any](name string) *Topic[T] {
	return &Topic[T]{
		name:          name,
		subscriptions: make(map[string]*Subscription[T]),
	}
}

func (t *Topic[T]) Name() string {
	return t.name
}

func (t *Topic[T]) Subscriptions() map[string]*Subscription[T] {
	return t.subscriptions
}

func (t *Topic[T]) Publish(message T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, subscription := range t.subscriptions {
		select {
		case subscription.ch <- message:
		default:
		}
	}
}

func (t *Topic[T]) NewSubscription(name string, bufferSize int, concurrency int64) *Subscription[T] {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.subscriptions[name]; !ok {
		t.subscriptions[name] = &Subscription[T]{
			name:        name,
			topic:       t,
			ch:          make(chan T, bufferSize),
			concurrency: concurrency,
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

type Subscription[T any] struct {
	name        string
	topic       *Topic[T]
	ch          chan T
	concurrency int64
}

func (s *Subscription[T]) Name() string {
	return s.name
}

func (s *Subscription[T]) Topic() *Topic[T] {
	return s.topic
}

func (s *Subscription[T]) Register() {
	s.topic.Register(s)
}

func (s *Subscription[T]) Unregister() {
	s.topic.Unregister(s)
}

func (s *Subscription[T]) Subscribe(ctx context.Context, consumer func(T)) {
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(s.concurrency)
	s.Register()

	for {
		select {
		case message := <-s.ch:
			err := s.do(ctx, &wg, sem, consumer, message)
			if err != nil && err != context.Canceled {
				panic(err)
			}

		case <-ctx.Done():
			s.Unregister()
			log.Printf("closing subscription: %s\n", s.Name())
			cancelCtx := context.Background()
			for {
				select {
				case message := <-s.ch:
					err := s.do(cancelCtx, &wg, sem, consumer, message)
					if err != nil {
						panic(err)
					}

				default:
					wg.Wait()
					return
				}
			}
		}
	}
}

func (s *Subscription[T]) do(ctx context.Context, wg *sync.WaitGroup, sem *semaphore.Weighted, consumer func(T), message T) error {
	wg.Add(1)
	if err := sem.Acquire(ctx, 1); err == nil {
		go func() {
			defer wg.Done()
			defer sem.Release(1)
			consumer(message)
		}()
		return nil

	} else {
		s.ch <- message
		wg.Done()
		return err
	}

}
