package main

import (
	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Octops/agones-event-broadcaster/pkg/broadcaster"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"reflect"
	"sync"
)

var (
	masterURL  string
	kubeconfig string
	address    string
)

type GameServer struct {
	Name    string
	Labels  map[string]string
	Address string
	Port    int32
	Status  string
}

type HTTPBroker struct {
	mutex sync.Mutex
	Store map[string]*GameServer
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
	case "gameserver.events.deleted":
		gsAgones := message.(*v1.GameServer)
		key := fmt.Sprintf("%s/%s", gsAgones.Namespace, gsAgones.Name)
		return h.handleDeleted(key)
	case "gameserver.events.added":
		gsAgones := message.(*v1.GameServer)
		return h.handleAdded(gsAgones)
	case "gameserver.events.updated":
		return h.handleUpdated(message)
	}

	return nil
}

func (h *HTTPBroker) handleDeleted(key string) error {
	h.DeleteGameServer(key)
	logrus.Infof("gameserver deleted %s", key)
	return nil
}

func (h *HTTPBroker) handleAdded(gsAgones *v1.GameServer) error {
	if gsAgones.Status.State == v1.GameServerStateReady {
		gs := &GameServer{
			Name:    fmt.Sprintf("%s/%s", gsAgones.Namespace, gsAgones.Name),
			Labels:  gsAgones.Labels,
			Address: gsAgones.Status.Address,
			Port:    gsAgones.Status.Ports[0].Port,
			Status:  string(gsAgones.Status.State),
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
			Status:  string(gsAgones.Status.State),
		}

		return h.AddGameServer(gs)
	}

	return nil
}

func main() {
	flag.Parse()

	if kubeconfig == "" {
		kubeconfig = flag.Lookup("kubeconfig").Value.String()
	}

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logrus.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	broker := &HTTPBroker{
		Store: map[string]*GameServer{},
	}

	http.HandleFunc("/", broker.Handler)
	go func() {
		logrus.Infof("server listening at %s", address)
		logrus.Fatal(http.ListenAndServe(address, nil))
	}()

	gsBroadcaster, err := broadcaster.New(cfg, broker)
	if err != nil {
		logrus.WithError(err).Fatal("error creating broadcaster")
	}

	if err := gsBroadcaster.Start(); err != nil {
		logrus.WithError(err).Fatal("error starting broadcaster")
	}
}

func init() {
	if flag.Lookup("kubeconfig") == nil {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	}

	if flag.Lookup("master") == nil {
		flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	}

	flag.StringVar(&address, "addr", ":8000", "The address of the HTTP server.")
}
