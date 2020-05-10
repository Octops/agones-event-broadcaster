package events

var (
	GameServerEventAdded   GameServerEventType = "gameserver.events.added"
	GameServerEventUpdated GameServerEventType = "gameserver.events.updated"
	GameServerEventDeleted GameServerEventType = "gameserver.events.deleted"
)

type GameServerEventType string

type GameServerEvent struct {
	Source EventSource
	Type   GameServerEventType
	Message
}

func GameServerAdded(message Message) *GameServerEvent {
	return &GameServerEvent{
		Source:  EventSourceOnAdd,
		Type:    GameServerEventAdded,
		Message: message,
	}
}

func GameServerUpdated(message Message) *GameServerEvent {
	return &GameServerEvent{
		Source:  EventSourceOnUpdate,
		Type:    GameServerEventUpdated,
		Message: message,
	}
}

func GameServerDeleted(message Message) *GameServerEvent {
	return &GameServerEvent{
		Source:  EventSourceOnDelete,
		Type:    GameServerEventDeleted,
		Message: message,
	}
}

func (t *GameServerEvent) EventType() EventType {
	return EventType(t.Type)
}

func (t *GameServerEvent) EventSource() EventSource {
	return t.Source
}

func (t GameServerEventType) String() string {
	return string(t)
}
