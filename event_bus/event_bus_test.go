package event_bus

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()
	assert.NotNil(t, eb)
	assert.NotNil(t, eb.listeners)
	assert.Empty(t, eb.listeners)
}

func TestSubscribe(t *testing.T) {
	testCases := []struct {
		name     string
		topic    string
		listener Listener
		expected int
	}{
		{
			name:  "single listener",
			topic: "test-topic",
			listener: Listener{
				TopicName: "test-topic",
				Handler: func(ctx context.Context, event *Event) error {
					return nil
				},
			},
			expected: 1,
		},
		{
			name:  "multiple listeners same topic",
			topic: "test-topic",
			listener: Listener{
				TopicName: "test-topic",
				Handler: func(ctx context.Context, event *Event) error {
					return nil
				},
			},
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eb := NewEventBus()
			eb.Subscribe(tc.topic, tc.listener)
			if tc.name == "multiple listeners same topic" {
				eb.Subscribe(tc.topic, tc.listener) // Add second listener
			}

			assert.Len(t, eb.listeners[tc.topic], tc.expected)
			for i := 0; i < tc.expected; i++ {
				assert.Equal(t, tc.topic, eb.listeners[tc.topic][i].TopicName)
				assert.NotNil(t, eb.listeners[tc.topic][i].Handler)
			}
		})
	}
}

func TestPublish(t *testing.T) {
	testCases := []struct {
		name           string
		topic          string
		eventData      any
		setupListeners func(eb *EventBus)
		expectedCalls  int
		expectError    bool
	}{
		{
			name:      "no listeners",
			topic:     "test-topic",
			eventData: "test-data",
			setupListeners: func(eb *EventBus) {
				// No listeners
			},
			expectedCalls: 0,
			expectError:   false,
		},
		{
			name:      "single listener",
			topic:     "test-topic",
			eventData: "test-data",
			setupListeners: func(eb *EventBus) {
				listener := Listener{
					TopicName: "test-topic",
					Handler: func(ctx context.Context, event *Event) error {
						return nil
					},
				}
				eb.Subscribe("test-topic", listener)
			},
			expectedCalls: 1,
			expectError:   false,
		},
		{
			name:      "multiple listeners",
			topic:     "test-topic",
			eventData: "test-data",
			setupListeners: func(eb *EventBus) {
				listener1 := Listener{
					TopicName: "test-topic",
					Handler: func(ctx context.Context, event *Event) error {
						return nil
					},
				}
				listener2 := Listener{
					TopicName: "test-topic",
					Handler: func(ctx context.Context, event *Event) error {
						return nil
					},
				}
				eb.Subscribe("test-topic", listener1)
				eb.Subscribe("test-topic", listener2)
			},
			expectedCalls: 2,
			expectError:   false,
		},
		{
			name:      "handler error",
			topic:     "test-topic",
			eventData: "test-data",
			setupListeners: func(eb *EventBus) {
				listener := Listener{
					TopicName: "test-topic",
					Handler: func(ctx context.Context, event *Event) error {
						return errors.New("handler error")
					},
				}
				eb.Subscribe("test-topic", listener)
			},
			expectedCalls: 1,
			expectError:   false, // Publish should not return error even if handler fails
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eb := NewEventBus()
			var callCount int
			var mu sync.Mutex

			// Setup listeners with counter
			originalSetup := tc.setupListeners
			tc.setupListeners = func(eb *EventBus) {
				originalSetup(eb)
				// Replace handlers to count calls
				for topic, listeners := range eb.listeners {
					for i := range listeners {
						originalHandler := listeners[i].Handler
						eb.listeners[topic][i].Handler = func(ctx context.Context, event *Event) error {
							mu.Lock()
							callCount++
							mu.Unlock()
							return originalHandler(ctx, event)
						}
					}
				}
			}
			tc.setupListeners(eb)

			event := &Event{
				Topic: tc.topic,
				Data:  tc.eventData,
			}

			err := eb.Publish(context.Background(), event)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Wait for async handlers to complete
			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			assert.Equal(t, tc.expectedCalls, callCount)
			mu.Unlock()
		})
	}
}

func TestPublishContextCancellation(t *testing.T) {
	eb := NewEventBus()
	var handlerCalled bool
	var mu sync.Mutex

	listener := Listener{
		TopicName: "test-topic",
		Handler: func(ctx context.Context, event *Event) error {
			mu.Lock()
			defer mu.Unlock()
			handlerCalled = true
			return nil
		},
	}

	eb.Subscribe("test-topic", listener)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	event := &Event{Topic: "test-topic", Data: "test-data"}

	err := eb.Publish(ctx, event)
	assert.NoError(t, err)

	// Wait for async handler to complete
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.True(t, handlerCalled) // Handler should still be called even with cancelled context
}

func TestConcurrentSubscribeAndPublish(t *testing.T) {
	eb := NewEventBus()
	var counter int
	var mu sync.Mutex

	// Start multiple goroutines subscribing
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listener := Listener{
				TopicName: "concurrent-topic",
				Handler: func(ctx context.Context, event *Event) error {
					mu.Lock()
					defer mu.Unlock()
					counter++
					return nil
				},
			}
			eb.Subscribe("concurrent-topic", listener)
		}()
	}

	wg.Wait()

	// Publish event
	event := &Event{Topic: "concurrent-topic", Data: "concurrent-data"}
	err := eb.Publish(context.Background(), event)
	assert.NoError(t, err)

	// Wait for all handlers to complete
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 10, counter)
}

func TestEventBusImplementsPublisher(t *testing.T) {
	var eb Publisher = NewEventBus()
	assert.NotNil(t, eb)
}
