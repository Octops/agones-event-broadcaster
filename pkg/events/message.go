package events

// EventMessage is the data structure for messages that are resulting of reconcile events.
type EventMessage struct {
	Body interface{}
}

// Content extracts the body of the EventMessage
func (e *EventMessage) Content() interface{} {
	return e.Body
}
