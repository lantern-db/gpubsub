# gpubsub - A generic PubSub messaging -

## Generate Topic

```go
topic := gpubsub.NewTopic[int](topicName, concurrency, interval, ttl)
```

## Generate Subscription

```go
subscription := topic.NewSubscription("DummyConsumer")
```

## Consume messages with callback

```go
subscription.Subscribe(ctx, func (m *gpubsub.Message[int]) {
	// get the content of message which has type T
	message := m.Body()
	
	// some consumer process 
	
	// Ack if succeed
	m.Ack()
	
	// Nack if failed, retry later
	m.Nack()
})
```

## Publish a message
```go
topic.Publish(1)
```

## Brief Example

```go
ctx, cancel := context.WithCancel(context.Background())
topicName := "DummyData"
concurrency := int64(2)
interval := 30 * time.Second
ttl := 1 * time.Hour

topic := gpubsub.NewTopic[int](topicName, concurrency, interval, ttl)
subscription := topic.NewSubscription("DummyConsumer")

var wg sync.WaitGroup

wg.Add(1)
go func() {
  defer wg.Done()
  subscription.Subscribe(ctx, func(m gpubsub.Message[int]) {
    log.Printf("data: %d\n", m.Body())
    m.Ack()
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
```

It will show belows.

```

2022/04/10 13:59:26 closing subscription: DummyConsumer
2022/04/10 13:59:27 data: 0
2022/04/10 13:59:27 data: 1
2022/04/10 13:59:28 data: 4
2022/04/10 13:59:28 data: 3
2022/04/10 13:59:29 data: 5
2022/04/10 13:59:29 data: 6
2022/04/10 13:59:30 data: 7
2022/04/10 13:59:30 data: 8
2022/04/10 13:59:31 data: 9
2022/04/10 13:59:31 data: 2

```
