package pubsub

import (
	"github.com/Octops/gameserver-events-broadcaster/pkg/events"
)

type Config struct {
	ProjectID string
	// TODO: Enable topic settings per event type
	//OnAddTopicID    string
	//OnUpdateTopicID string
	//OnDeleteTopicID string
}

type PubSubBroker struct {
	*Config
}

func NewPubSubBroker(config *Config) *PubSubBroker {
	return &PubSubBroker{Config: config}
}

func (ps *PubSubBroker) BuildEnvelope(event events.Event) (*events.Envelope, error) {
	envelope := &events.Envelope{}
	envelope.AddHeader("event_type", event.EventType())
	envelope.AddHeader("project_id", ps.ProjectID)
	envelope.AddHeader("topic_id", "gameserver.events")

	envelope.Message = event.(events.Message).Content()

	return envelope, nil
}

func (ps *PubSubBroker) SendMessage(envelope *events.Envelope) error {
	panic("implement me")
}
