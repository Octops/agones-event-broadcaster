package events

var (
	EventSourceOnAdd    EventSource = "OnAdd"
	EventSourceOnUpdate EventSource = "OnUpdate"
	EventSourceOnDelete EventSource = "OnDelete"
)

type EventSource string
type EventType string

type Event interface {
	EventType() EventType
	EventSource() EventSource
}
type Message interface {
	Content() interface{}
}

func (t EventType) String() string {
	return string(t)
}
