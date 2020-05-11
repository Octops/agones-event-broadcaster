package events

var (
	EventSourceOnAdd    EventSource = "OnAdd"
	EventSourceOnUpdate EventSource = "OnUpdate"
	EventSourceOnDelete EventSource = "OnDelete"
)

type EventSource string
type EventType string

// Event is the contract for events handled by EventHandlers
type Event interface {
	EventType() EventType
	EventSource() EventSource
}

// Message is the contract for messages published by Brokers
type Message interface {
	Content() interface{}
}

// String returns the string representation of a EventType
func (t EventType) String() string {
	return string(t)
}
