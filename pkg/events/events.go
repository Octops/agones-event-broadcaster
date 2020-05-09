package events

type EventMessage struct {
	Body interface{}
}

func (g *EventMessage) Content() interface{} {
	return g.Body
}

type GameServerDeleted struct {
	Event
	Message
}

func (g *GameServerDeleted) EventType() string {
	return "GameServerDeleted"
}
