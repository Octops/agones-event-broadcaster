package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/Octops/gameserver-events-broadcaster/pkg/events"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

const (
	PROJECTID_HEADER_KEY  = "pubsub_project_id"
	TOPIC_ID_HEADER_KEY   = "pubsub_topic_id"
	EVENT_TYPE_HEADER_KEY = "pubsub_event_type"
	DEFAULT_TOPIC_ID      = "gameserver.events"
)

// GenericTopicID is used when specific events topics are not set. Defaults to "gameserver.events"
type Config struct {
	ProjectID       string
	GenericTopicID  string
	OnAddTopicID    string
	OnUpdateTopicID string
	OnDeleteTopicID string
}

type PubSubBroker struct {
	*Config
	*pubsub.Client
}

func NewPubSubBroker(config *Config, opts ...option.ClientOption) (*PubSubBroker, error) {
	config.ApplyDefaults()

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.ProjectID, opts...)
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

	b.SetEnvelopeHeader(event, envelope)

	envelope.Message = event.(events.Message).Content()

	return envelope, nil
}

func (b *PubSubBroker) SetEnvelopeHeader(event events.Event, envelope *events.Envelope) {
	var topicID string

	switch event.EventSource() {
	case events.EventSourceOnAdd:
		topicID = b.OnAddTopicID
	case events.EventSourceOnUpdate:
		topicID = b.OnUpdateTopicID
	case events.EventSourceOnDelete:
		topicID = b.OnDeleteTopicID
	default:
		topicID = b.GenericTopicID
	}

	envelope.AddHeader(TOPIC_ID_HEADER_KEY, topicID)
	envelope.AddHeader(EVENT_TYPE_HEADER_KEY, event.EventType().String())
	envelope.AddHeader(PROJECTID_HEADER_KEY, b.ProjectID)
}

func (b *PubSubBroker) SendMessage(envelope *events.Envelope) error {
	ctx := context.Background()

	topicID, ok := GetTopicIDFromHeader(envelope)
	if !ok {
		return fmt.Errorf("topicID is not present on the envelope header")
	}

	messageID, err := b.publish(ctx, envelope, topicID)
	if err != nil {
		logrus.WithError(err).Errorf("error publishing message to topic %s", topicID)
		return err
	}

	logrus.WithField("broker", "pubsub").Infof("message published to topicID:\"%s\" messageID:\"%s\"", topicID, messageID)

	return nil
}

func (b *PubSubBroker) TopicFor(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	topic := b.Client.Topic(topicID)

	// TODO: This check requires Pub/Sub Editor role.
	// Review if checking if topic exists is worth having such a role
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not check if topic exists: %v", err)
	}

	if !ok {
		return nil, fmt.Errorf("topic %s for projectID %s does not exist", b.ProjectID, topicID)
	}

	return topic, err
}

func (b *PubSubBroker) publish(ctx context.Context, envelope *events.Envelope, topicID string) (string, error) {
	msg, err := envelope.Encode()
	if err != nil {
		return "", fmt.Errorf("error encoding envelope: %v", err)
	}

	topic, err := b.TopicFor(ctx, topicID)
	if err != nil {
		return "", fmt.Errorf("error building topic %s: %v", topicID, err)
	}

	// TODO: Implement Publish in batches
	result := topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting result for the message published to topic \"%s\": %v", topicID, err)
	}

	return id, nil
}

func (c *Config) ApplyDefaults() {
	c.GenericTopicID = CheckEmpty(c.GenericTopicID, DEFAULT_TOPIC_ID)
	c.OnAddTopicID = CheckEmpty(c.OnAddTopicID, DEFAULT_TOPIC_ID)
	c.OnUpdateTopicID = CheckEmpty(c.OnUpdateTopicID, DEFAULT_TOPIC_ID)
	c.OnDeleteTopicID = CheckEmpty(c.OnDeleteTopicID, DEFAULT_TOPIC_ID)
}

func GetTopicIDFromHeader(envelope *events.Envelope) (string, bool) {
	if topicID, ok := envelope.Header.Headers[TOPIC_ID_HEADER_KEY]; ok {
		return topicID, true
	}

	return "", false
}

func CheckEmpty(source, newValue string) string {
	if source == "" {
		return newValue
	}
	return source
}
