package main

import (
	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"encoding/json"
	"fmt"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"sync"
)

type GameServer struct {
	Name    string            `json:"name"`
	Labels  map[string]string `json:"labels"`
	Address string            `json:"address"`
	Port    int32             `json:"port"`
	State   string            `json:"state"`
}

type HTTPBroker struct {
	mutex sync.Mutex
	Store map[string]*GameServer
}

func NewHTTPBroker(addr string) *HTTPBroker {
	return &HTTPBroker{
		Store: map[string]*GameServer{},
	}
}

func (h *HTTPBroker) BuildEnvelope(event events.Event) (*events.Envelope, error) {
	envelope := &events.Envelope{}

	envelope.AddHeader("event_type", event.EventType().String())
	envelope.Message = event.(events.Message)

	return envelope, nil
}

func (h *HTTPBroker) SendMessage(envelope *events.Envelope) error {
	message := envelope.Message.(events.Message).Content()
	eventType := envelope.Header.Headers["event_type"]

	switch eventType {
	case "gameserver.events.added":
		gsAgones := message.(*v1.GameServer)
		return h.handleAdded(gsAgones)
	case "gameserver.events.updated":
		return h.handleUpdated(message)
	case "gameserver.events.deleted":
		gsAgones := message.(*v1.GameServer)
		return h.handleDeleted(gsAgones)
	}

	return nil
}

func (h *HTTPBroker) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.ListGameServer())
}

func (h *HTTPBroker) AddGameServer(gs *GameServer) error {
	defer h.mutex.Unlock()

	h.mutex.Lock()

	h.Store[gs.Name] = gs
	return nil
}

func (h *HTTPBroker) DeleteGameServer(key string) {
	defer h.mutex.Unlock()

	h.mutex.Lock()

	delete(h.Store, key)
}

func (h *HTTPBroker) ListGameServer() map[string]*GameServer {
	defer h.mutex.Unlock()

	h.mutex.Lock()

	return h.Store
}

func (h *HTTPBroker) handleAdded(gsAgones *v1.GameServer) error {
	if gsAgones.Status.State == v1.GameServerStateReady {
		gs := &GameServer{
			Name:    fmt.Sprintf("%s/%s", gsAgones.Namespace, gsAgones.Name),
			Labels:  gsAgones.Labels,
			Address: gsAgones.Status.Address,
			Port:    gsAgones.Status.Ports[0].Port,
			State:   string(gsAgones.Status.State),
		}

		return h.AddGameServer(gs)
	}

	return nil
}

func (h *HTTPBroker) handleUpdated(message interface{}) error {
	msgUpdate := reflect.ValueOf(message)
	gsAgones := msgUpdate.Field(1).Interface().(*v1.GameServer)

	if gsAgones.Status.State == v1.GameServerStateReady {
		gs := &GameServer{
			Name:    fmt.Sprintf("%s/%s", gsAgones.Namespace, gsAgones.Name),
			Labels:  gsAgones.Labels,
			Address: gsAgones.Status.Address,
			Port:    gsAgones.Status.Ports[0].Port,
			State:   string(gsAgones.Status.State),
		}

		return h.AddGameServer(gs)
	}

	return nil
}

func (h *HTTPBroker) handleDeleted(gsAgones *v1.GameServer) error {
	key := fmt.Sprintf("%s/%s", gsAgones.Namespace, gsAgones.Name)
	h.DeleteGameServer(key)
	logrus.Infof("gameserver deleted %s", key)
	return nil
}
