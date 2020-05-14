package brokers

import "github.com/Octops/agones-event-broadcaster/pkg/events"

// Broker is the service used by the Broadcaster for publishing events
type Broker interface {
	BuildEnvelope(event events.Event) (*events.Envelope, error)
	SendMessage(envelope *events.Envelope) error
}
