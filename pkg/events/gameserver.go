package events

import v1 "agones.dev/agones/pkg/apis/agones/v1"

var (
	GameServerEventAdded   GameServerEventType = "gameserver.events.added"
	GameServerEventUpdated GameServerEventType = "gameserver.events.updated"
	GameServerEventDeleted GameServerEventType = "gameserver.events.deleted"
)

type GameServerEventType string

// GameServerEvent is the data structure for reconcile events associated with Agones GameServers
// It holds the event source (OnAdd, OnUpdate, OnDelete) and the event type (Added, Updated, Deleted).
type GameServerEvent struct {
	Source  EventSource         `json:"source"`
	Type    GameServerEventType `json:"type"`
	Message `json:"message"`
}

func init() {
	RegisterEventFactory(&v1.GameServer{}, GameServerAdded, GameServerUpdated, GameServerDeleted)
}

// GameServerAdded is the data structure for reconcile events of type Add
func GameServerAdded(message Message) Event {
	return &GameServerEvent{
		Source:  EventSourceOnAdd,
		Type:    GameServerEventAdded,
		Message: message,
	}
}

// GameServerUpdates is the data structure for reconcile events of type Update
func GameServerUpdated(message Message) Event {
	return &GameServerEvent{
		Source:  EventSourceOnUpdate,
		Type:    GameServerEventUpdated,
		Message: message,
	}
}

// GameServerDeleted is the data structure for reconcile events of type Delete
func GameServerDeleted(message Message) Event {
	return &GameServerEvent{
		Source:  EventSourceOnDelete,
		Type:    GameServerEventDeleted,
		Message: message,
	}
}

// EventType returns the type of the reconcile event for a GameServer.
// For example: Added, Updated, Deleted
func (t *GameServerEvent) EventType() EventType {
	return EventType(t.Type)
}

// EventSource return the event source that generated the event.
// For example: OnAdd, OnUpdate, OnDelete
func (t *GameServerEvent) EventSource() EventSource {
	return t.Source
}

// String is a helper method that returns the string version of a GameServerEventType
func (t GameServerEventType) String() string {
	return string(t)
}
