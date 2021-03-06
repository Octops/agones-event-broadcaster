package events

import (
	v1 "agones.dev/agones/pkg/apis/agones/v1"
)

var (
	FleetEventAdded   FleetEventType = "fleet.events.added"
	FleetEventUpdated FleetEventType = "fleet.events.updated"
	FleetEventDeleted FleetEventType = "fleet.events.deleted"
)

type FleetEventType string

// FleetEvent is the data structure for reconcile events associated with Agones Fleets
// It holds the event source (OnAdd, OnUpdate, OnDelete) and the event type (Added, Updated, Deleted).
type FleetEvent struct {
	Source  EventSource    `json:"source"`
	Type    FleetEventType `json:"type"`
	Message `json:"message"`
}

func init() {
	RegisterEventFactory(&v1.Fleet{}, FleetAdded, FleetUpdated, FleetDeleted)
}

// FleetAdded is the data structure for reconcile events of type Add
func FleetAdded(message Message) Event {
	return &FleetEvent{
		Source:  EventSourceOnAdd,
		Type:    FleetEventAdded,
		Message: message,
	}
}

// FleetUpdated is the data structure for reconcile events of type Update
func FleetUpdated(message Message) Event {
	return &FleetEvent{
		Source:  EventSourceOnUpdate,
		Type:    FleetEventUpdated,
		Message: message,
	}
}

// FleetDeleted is the data structure for reconcile events of type Delete
func FleetDeleted(message Message) Event {
	return &FleetEvent{
		Source:  EventSourceOnDelete,
		Type:    FleetEventDeleted,
		Message: message,
	}
}

// EventType returns the type of the reconcile event for a Fleet.
// For example: Added, Updated, Deleted
func (t *FleetEvent) EventType() EventType {
	return EventType(t.Type)
}

// EventSource return the event source that generated the event.
// For example: OnAdd, OnUpdate, OnDelete
func (t *FleetEvent) EventSource() EventSource {
	return t.Source
}

// String is a helper method that returns the string version of a FleetEventType
func (t FleetEventType) String() string {
	return string(t)
}
