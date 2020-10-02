package events

import (
	"reflect"
)

var (
	AddedEventsRegistry   = map[string]func(message Message) Event{}
	UpdatedEventsRegistry = map[string]func(message Message) Event{}
	DeletedEventsRegistry = map[string]func(message Message) Event{}
)

func ForAdded(message Message) Event {
	c := message.Content()
	kind := reflect.TypeOf(c).Elem().String()
	fn, ok := AddedEventsRegistry[kind]
	if !ok {
		return nil
	}

	return fn(message)
}

func ForUpdated(message Message) Event {
	c := message.Content()
	m := reflect.ValueOf(c)
	obj := m.Field(1).Interface()

	kind := reflect.TypeOf(obj).Elem().String()
	fn, ok := UpdatedEventsRegistry[kind]
	if !ok {
		return nil
	}

	return fn(message)
}

func ForDeleted(message Message) Event {
	c := message.Content()
	kind := reflect.TypeOf(c).Elem().String()
	fn, ok := DeletedEventsRegistry[kind]
	if !ok {
		return nil
	}

	return fn(message)
}
