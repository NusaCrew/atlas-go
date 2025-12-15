package event_bus

import (
	"context"
	"sync"

	"github.com/NusaCrew/atlas-go/log"
)

type Event struct {
	Topic    string
	Data     any
	Metadata map[string]any
}

type Handler func(ctx context.Context, event *Event) error

type Listener struct {
	TopicName      string
	SubscriberName string
	Handler        Handler
}

type EventBus struct {
	mu        sync.RWMutex
	listeners map[string][]Listener
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]Listener),
	}
}

func (eb *EventBus) Subscribe(topic string, listener Listener) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.listeners[topic] = append(eb.listeners[topic], listener)
}

func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	eb.mu.RLock()
	listeners := eb.listeners[event.Topic]
	eb.mu.RUnlock()

	for _, listener := range listeners {
		go func(l Listener) {
			logHandler(l, ctx, event)
		}(listener)
	}
	return nil
}

func logHandler(l Listener, ctx context.Context, event *Event) {
	tracer := log.NewTracer(ctx, l.SubscriberName, "EventBus")

	tracer.WithFields(
		map[string]any{
			"subscriber": l.SubscriberName,
			"topic":      l.TopicName,
		},
	).Info(log.OK, log.Request, "starting event handler for topic %s", event.Topic)

	err := l.Handler(ctx, event)
	if err != nil {
		// TODO: trigger alert here...
	}
	defer tracer.TraceResponse(err)
}
