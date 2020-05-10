package events

type GameServerEventType string

var (
	GameServerEventAdded   = "gameserver.events.added"
	GameServerEventUpdated = "gameserver.events.updated"
	GameServerEventDeleted = "gameserver.events.deleted"
)

type GameServerEvent struct {
	Type string
	Message
}

func GameServerDeleted(message Message) *GameServerEvent {
	return &GameServerEvent{
		Type:    GameServerEventDeleted,
		Message: message,
	}
}

func GameServerAdded(message Message) *GameServerEvent {
	return &GameServerEvent{
		Type:    GameServerEventAdded,
		Message: message,
	}
}

func GameServerUpdated(message Message) *GameServerEvent {
	return &GameServerEvent{
		Type:    GameServerEventUpdated,
		Message: message,
	}
}

func (t *GameServerEvent) EventType() string {
	return t.Type
}
