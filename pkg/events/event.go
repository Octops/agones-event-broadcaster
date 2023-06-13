package events

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
)

var (
	EventFactoryRegistry = map[string]*EventFactory{}
)

type EventFactory struct {
	OnAdded   EventBuilder
	OnUpdated EventBuilder
	OnDeleted EventBuilder
}

type EventBuilder func(message Message) Event

// RegisterEventFactory register events builders for a particular resource type.
func RegisterEventFactory(obj runtime.Object, onAdded EventBuilder, onUpdated EventBuilder, onDeleted EventBuilder) {
	kind := reflect.TypeOf(obj).Elem().String()
	EventFactoryRegistry[kind] = &EventFactory{
		OnAdded:   onAdded,
		OnUpdated: onUpdated,
		OnDeleted: onDeleted,
	}
}

// OnAdded builds an event of type OnAdded for a particular message content type
func OnAdded(message Message) Event {
	c := message.Content()
	kind := ResourceMessageKind(c.(runtime.Object))

	fn, ok := EventFactoryRegistry[kind]
	if !ok {
		return nil
	}

	return fn.OnAdded(message)
}

// OnUpdated builds an event of type OnUpdated for a particular message content type
func OnUpdated(message Message) Event {
	c := message.Content()
	m := reflect.ValueOf(c)
	obj := m.Field(1).Interface()

	kind := ResourceMessageKind(obj.(runtime.Object))
	fn, ok := EventFactoryRegistry[kind]
	if !ok {
		return nil
	}

	return fn.OnUpdated(message)
}

// OnDeleted builds an event of type OnDeleted for a particular message content type
func OnDeleted(message Message) Event {
	c := message.Content()
	kind := ResourceMessageKind(c.(runtime.Object))

	fn, ok := EventFactoryRegistry[kind]
	if !ok {
		return nil
	}

	return fn.OnDeleted(message)
}

// ResourceMessageKind returns the type of the object that is the content of the message
func ResourceMessageKind(obj runtime.Object) string {
	return reflect.TypeOf(obj).Elem().String()
}
