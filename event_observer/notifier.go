package event_observer

import "context"

type EventNotifier interface {
	NotifySubscribers(ctx context.Context, event *Event)
}
