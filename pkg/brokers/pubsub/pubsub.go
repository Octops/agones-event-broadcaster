package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
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

	// TODO: Implement Options https://pkg.go.dev/google.golang.org/api/option@v0.13.0?tab=doc#ClientOption
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

	topic, err := b.TopicFor(ctx, topicID)
	if err != nil {
		return fmt.Errorf("error building topic %s: %v", topicID, err)
	}

	msg, err := envelope.Encode()
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
		return fmt.Errorf("error getting result for the message published to topic \"%s\": %v", topicID, err)
	}

	logrus.WithField("broker", "pubsub").Infof("message published to topicID:\"%s\" messageID:\"%s\"", topicID, id)

	return nil
}

func (b *PubSubBroker) TopicFor(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	topic := b.Client.Topic(topicID)

	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not check if topic exists: %v", err)
	}

	if !ok {
		return nil, fmt.Errorf("topic %s for projectID %s does not exist", b.ProjectID, topicID)
	}

	return topic, err
}

func GetTopicIDFromHeader(envelope *events.Envelope) (string, bool) {
	if topicID, ok := envelope.Header.Headers["topic_id"]; ok {
		return topicID, true
	}

	return "", false
}
