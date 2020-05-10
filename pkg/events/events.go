package events

type EventMessage struct {
	Body interface{}
}

func (e *EventMessage) Content() interface{} {
	return e.Body
}

type GameServerDeleted struct {
	Event
	Message
}

type GameServerAdded struct {
	Event
	Message
}

func (g *GameServerAdded) EventType() string {
	return "gameserver.events.added"
}

func (g *GameServerDeleted) EventType() string {
	return "gameserver.events.deleted"
}
