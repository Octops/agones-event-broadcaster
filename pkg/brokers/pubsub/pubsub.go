package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Octops/gameserver-events-broadcaster/pkg/events"
	"github.com/sirupsen/logrus"
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
	*pubsub.Client
}

func NewPubSubBroker(config *Config) (*PubSubBroker, error) {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("error creating pubsub client for projectID %s: %v", config.ProjectID, err)
	}

	return &PubSubBroker{
		Config: config,
		Client: client,
	}, nil
}

func (b *PubSubBroker) BuildEnvelope(event events.Event) (*events.Envelope, error) {
	envelope := &events.Envelope{}
	envelope.AddHeader("event_type", event.EventType())
	envelope.AddHeader("project_id", b.ProjectID)
	envelope.AddHeader("topic_id", "gameserver.events")

	envelope.Message = event.(events.Message).Content()

	return envelope, nil
}

func (b *PubSubBroker) SendMessage(envelope *events.Envelope) error {
	ctx := context.Background()

	topicID, ok := GetTopicIDFromHeader(envelope)
	if !ok {
		return fmt.Errorf("topicID is not present on the envelope header")
	}

	topic := b.Client.Topic(topicID)

	ok, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("could not check if topic exists: %v", err)
	}

	if !ok {
		return fmt.Errorf("topic %s for projectID %s does not exist", b.ProjectID, topicID)
	}

	msg, err := EncodedEnvelope(envelope)
	if err != nil {
		return fmt.Errorf("error encoding envelope: %v", err)
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}

	logrus.WithField("broker", "pubsub").Infof("message published topicID:\"%s\" messageID:\"%s\"", topicID, id)

	return nil
}

func GetTopicIDFromHeader(envelope *events.Envelope) (string, bool) {
	if topicID, ok := envelope.Header.Headers["topic_id"]; ok {
		return topicID, true
	}

	return "", false
}

func EncodedEnvelope(envelope *events.Envelope) ([]byte, error) {
	return json.Marshal(envelope)
}
