package brokers

import "github.com/Octops/gameserver-events-broadcaster/pkg/events"

// Business logic, build envelope, parse message, call Dispatcher
type Broker interface {
	BuildEnvelope(event events.Event) (*events.Envelope, error)
	SendMessage(envelope *events.Envelope) error
}
