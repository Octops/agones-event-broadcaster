package events

type Event interface {
	EventType() string
}
type Message interface {
	Content() interface{}
}
