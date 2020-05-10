package events

type GameServerEvent struct {
	Type string
	Message
}

func GameServerDeleted(message Message) *GameServerEvent {
	return &GameServerEvent{
		Type:    "gameserver.events.deleted",
		Message: message,
	}
}

func GameServerAdded(message Message) *GameServerEvent {
	return &GameServerEvent{
		Type:    "gameserver.events.added",
		Message: message,
	}
}

func GameServerUpdated(message Message) *GameServerEvent {
	return &GameServerEvent{
		Type:    "gameserver.events.updated",
		Message: message,
	}
}

func (t *GameServerEvent) EventType() string {
	return t.Type
}
