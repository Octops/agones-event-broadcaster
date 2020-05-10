package events

type EventMessage struct {
	Body interface{}
}

func (e *EventMessage) Content() interface{} {
	return e.Body
}
