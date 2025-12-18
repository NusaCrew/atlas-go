package event_observer

import (
	"context"
	"fmt"
	"sync"

	"github.com/NusaCrew/atlas-go/log"
)

type Event struct {
	Topic    string
	Data     any
	Metadata map[string]any
}

type HandlerFunc func(ctx context.Context, event *Event) error

type Subscriber struct {
	TopicName      string
	SubscriberName string
	HandlerFunc    HandlerFunc
}

type EventObserver struct {
	mu          sync.RWMutex
	serviceName string
	subscribers map[string][]Subscriber
}

func NewEventObserver(serviceName string) *EventObserver {
	return &EventObserver{
		serviceName: serviceName,
		subscribers: make(map[string][]Subscriber),
	}
}

func (eo *EventObserver) Subscribe(topic string, subscriber Subscriber) {
	eo.mu.Lock()
	defer eo.mu.Unlock()
	eo.subscribers[topic] = append(eo.subscribers[topic], subscriber)
	log.Info("subscriber %s successfully joined topic %s", subscriber.SubscriberName, topic)
}

func (eo *EventObserver) NotifySubscribers(ctx context.Context, event *Event) {
	eo.mu.RLock()
	subscribers := eo.subscribers[event.Topic]
	eo.mu.RUnlock()

	for _, subscriber := range subscribers {
		go func(s Subscriber) {
			tracer := log.NewTracer(ctx, s.SubscriberName, fmt.Sprintf("EventObserver-%s", eo.serviceName)).WithFields(map[string]any{
				"subscriber": s.SubscriberName,
				"topic":      s.TopicName,
			})

			tracer.Info(log.OK, log.Request, "starting event handler for topic %s", event.Topic)

			err := s.HandlerFunc(ctx, event)
			if err != nil {
				tracer.WithField("operation", "subscription_handler").Error(log.ServerError, log.Response, err, "got error from subscription %s with topic %s", s.SubscriberName, s.TopicName)
			}
			defer tracer.TraceResponse(err)
		}(subscriber)
	}

}
