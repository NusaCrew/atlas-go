package event_observer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEventObserver(t *testing.T) {
	eo := NewEventObserver("some-service-name")
	assert.NotNil(t, eo)
	assert.NotNil(t, eo.subscribers)
	assert.Equal(t, "some-service-name", eo.serviceName)
	assert.Empty(t, eo.subscribers)
}

func TestSubscribe(t *testing.T) {
	eo := NewEventObserver("some-service-name")
	subscriber := Subscriber{
		TopicName: "test-topic",
		HandlerFunc: func(ctx context.Context, event *Event) error {
			return nil
		},
	}
	eo.Subscribe("test-topic", subscriber)

	assert.Len(t, eo.subscribers["test-topic"], 1)
	assert.Equal(t, "test-topic", eo.subscribers["test-topic"][0].TopicName)
}

func TestNotifySubscribers(t *testing.T) {
	eo := NewEventObserver("some-service-name")
	var callCount int
	var mu sync.Mutex

	subscriber := Subscriber{
		TopicName: "test-topic",
		HandlerFunc: func(ctx context.Context, event *Event) error {
			mu.Lock()
			callCount++
			mu.Unlock()
			return nil
		},
	}
	eo.Subscribe("test-topic", subscriber)

	e := &Event{Topic: "test-topic", Data: "payload"}
	eo.NotifySubscribers(context.Background(), e)

	// Wait for goroutine to run
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 1, callCount)
	mu.Unlock()
}

func TestNotifySubscribersHandlerError(t *testing.T) {
	eo := NewEventObserver("some-service-name")
	var called bool
	var mu sync.Mutex

	subscriber := Subscriber{
		TopicName: "test-topic",
		HandlerFunc: func(ctx context.Context, event *Event) error {
			mu.Lock()
			called = true
			mu.Unlock()
			return assert.AnError
		},
	}
	eo.Subscribe("test-topic", subscriber)

	e := &Event{Topic: "test-topic", Data: "payload"}
	eo.NotifySubscribers(context.Background(), e)

	// Wait for goroutine to run
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	assert.True(t, called)
	mu.Unlock()
}

func TestNotifyContextCancellation(t *testing.T) {
	eo := NewEventObserver("some-service-name")
	var called bool
	var mu sync.Mutex

	subscriber := Subscriber{
		TopicName: "test-topic",
		HandlerFunc: func(ctx context.Context, event *Event) error {
			mu.Lock()
			called = true
			mu.Unlock()
			return nil
		},
	}
	eo.Subscribe("test-topic", subscriber)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	e := &Event{Topic: "test-topic", Data: "payload"}
	eo.NotifySubscribers(ctx, e)

	// Handler should still run even if context already cancelled
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	assert.True(t, called)
	mu.Unlock()
}

func TestConcurrentSubscribeAndNotify(t *testing.T) {
	eo := NewEventObserver("some-service-name")
	var counter int
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subscriber := Subscriber{
				TopicName: "concurrent-topic",
				HandlerFunc: func(ctx context.Context, event *Event) error {
					mu.Lock()
					counter++
					mu.Unlock()
					return nil
				},
			}
			eo.Subscribe("concurrent-topic", subscriber)
		}()
	}

	wg.Wait()

	e := &Event{Topic: "concurrent-topic", Data: "payload"}
	eo.NotifySubscribers(context.Background(), e)

	// Wait for handlers
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 10, counter)
	mu.Unlock()
}
