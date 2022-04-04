# gpubsub - A generic PubSub messaging -

## Generate Topic

```go
topic := gpubsub.NewTopic[int]("DummyData")
```

## Generate Subscription

```go
subscription := topic.NewSubscription("DummyConsumer", 10000)
```

## Consume messages with callback

```go
subscription.Subscribe(ctx, func (m int) {
  // some consumer process
})
```

## Publish a message
```go
topic.Publish(1)
```

## Brief Example

```go
ctx, cancel := context.WithCancel(context.Background())

topic := gpubsub.NewTopic[int]("DummyData")
subscription := topic.NewSubscription("DummyConsumer", 10000, 2)

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
```

It will show belows.

```
2022/03/25 21:49:09 data: 0
2022/03/25 21:49:09 data: 1
2022/03/25 21:49:09 data: 2
2022/03/25 21:49:09 data: 3
2022/03/25 21:49:09 data: 4
2022/03/25 21:49:09 data: 5
2022/03/25 21:49:09 data: 6
2022/03/25 21:49:09 data: 7
2022/03/25 21:49:09 data: 8
2022/03/25 21:49:09 closing subscription: DummyConsumer
2022/03/25 21:49:09 data: 9

```
