package gpubsub

import (
	"context"
	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
	"log"
	"sync"
	"time"
)

type Subscription[T any] struct {
	mu          sync.RWMutex
	name        string
	topic       *Topic[T]
	ch          chan string
	messages    map[string]*Message[T]
	concurrency int64
	interval    time.Duration
	ttl         time.Duration
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

func (s *Subscription[T]) NewMessage(body T) *Message[T] {
	return &Message[T]{
		id:           uuid.New().String(),
		body:         body,
		subscription: s,
		createdAt:    time.Now(),
	}
}

func (s *Subscription[T]) Publish(message *Message[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ch <- message.id
	s.messages[message.id] = message
}

func (s *Subscription[T]) ack(message *Message[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.messages, message.id)
}

func (s *Subscription[T]) nack(message *Message[T]) {
	message.touch()
	s.messages[message.id] = message
}

func (s *Subscription[T]) remind(message *Message[T]) {
	s.ch <- message.id
}

func (s *Subscription[T]) Subscribe(ctx context.Context, consumer func(*Message[T])) {
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(s.concurrency)
	s.Register()
	go s.watch(ctx, s.interval, s.ttl)

	for {
		select {
		case id := <-s.ch:
			message := s.messages[id]
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
				case id := <-s.ch:
					message := s.messages[id]
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

func (s *Subscription[T]) do(ctx context.Context, wg *sync.WaitGroup, sem *semaphore.Weighted, consumer func(*Message[T]), message *Message[T]) error {
	wg.Add(1)
	if err := sem.Acquire(ctx, 1); err == nil {
		go func() {
			defer wg.Done()
			defer sem.Release(1)
			defer message.touch()
			consumer(message)
		}()
		return nil

	} else {
		s.remind(message)
		wg.Done()
		return err
	}
}

func (s *Subscription[T]) salvage(interval time.Duration, ttl time.Duration) {
	for _, message := range s.messages {
		if time.Now().Sub(message.createdAt) > ttl {
			s.ack(message)
		}

		if time.Now().Sub(message.lastViewedAt) > interval {
			s.remind(message)
		}
	}
}

func (s *Subscription[T]) watch(ctx context.Context, interval time.Duration, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			s.salvage(interval, ttl)
		case <-ctx.Done():
			return
		}
	}
}
