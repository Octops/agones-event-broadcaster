package handlers

// EventHandler defines the contract for event handlers used by the GameSever controller
// when notifying reconcile events
type EventHandler interface {
	OnAdd(obj interface{}) error
	OnUpdate(oldObj interface{}, newObj interface{}) error
	OnDelete(obj interface{}) error
}
