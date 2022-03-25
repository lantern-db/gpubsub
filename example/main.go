package main

import (
	"context"
	"gpubsub/gpubsub"
	"log"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	topic := gpubsub.NewTopic[int]("DummyData")
	subscription := topic.NewSubscription("DummyConsumer", 10000)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		subscription.Subscribe(ctx, func(m int) {
			log.Printf("data: %d\n", m)
		})
	}()

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Millisecond)
		topic.Publish(i)
	}
	cancel()
	for j := 11; j < 20; j++ {
		time.Sleep(1 * time.Millisecond)
		topic.Publish(j)
	}

	wg.Wait()
}
