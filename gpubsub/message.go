package gpubsub

import "time"

type Message[T any] struct {
	messageID    string
	body         T
	subscription Subscription[T]
	createdAt    time.Time
}
