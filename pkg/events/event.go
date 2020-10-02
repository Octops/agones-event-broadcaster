package events

import (
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
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

func RegisterEventFactory(obj runtime.Object, onAdded EventBuilder, onUpdated EventBuilder, onDeleted EventBuilder) {
	kind := reflect.TypeOf(obj).Elem().String()
	EventFactoryRegistry[kind] = &EventFactory{
		OnAdded:   onAdded,
		OnUpdated: onUpdated,
		OnDeleted: onDeleted,
	}
}

func ForAdded(message Message) Event {
	c := message.Content()
	kind := ResourceMessageKind(c.(runtime.Object))

	fn, ok := EventFactoryRegistry[kind]
	if !ok {
		return nil
	}

	return fn.OnAdded(message)
}

func ForUpdated(message Message) Event {
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

func ForDeleted(message Message) Event {
	c := message.Content()
	kind := ResourceMessageKind(c.(runtime.Object))

	fn, ok := EventFactoryRegistry[kind]
	if !ok {
		return nil
	}

	return fn.OnDeleted(message)
}

func ResourceMessageKind(obj runtime.Object) string {
	return reflect.TypeOf(obj).Elem().String()
}
