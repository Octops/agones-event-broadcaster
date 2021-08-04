package kafka

import (
	"fmt"

	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type KafkaBroker struct {
	*Config
	*kafka.Producer
	*kafka.AdminClient
}

type Config struct {
	GenericTopicID   string
	OnAddTopicID     string
	OnUpdateTopicID  string
	OnDeleteTopicID  string
	APIKey           string
	APISecret        string
	BootstrapServers string
}

const (
	TOPIC_ID_HEADER_KEY   = "kafka_topic_id"
	EVENT_TYPE_HEADER_KEY = "kafka_event_type"
	DEFAULT_TOPIC_ID      = "gameserver.events"
)

func NewKafkaBroker(config *Config) (*KafkaBroker, error) {
	config.ApplyDefaults()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServers,
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"sasl.username":     config.APIKey,
		"sasl.password":     config.APISecret})

	if err != nil {
		return nil, fmt.Errorf("failed to create Producer client: %s\n", err)
	}

	return &KafkaBroker{
		Config:   config,
		Producer: producer,
	}, nil
}
func (k *KafkaBroker) BuildEnvelope(event events.Event) (*events.Envelope, error) {
	envelope := &events.Envelope{}

	k.SetEnvelopeHeader(event, envelope)

	envelope.Message = event.(events.Message).Content()

	return envelope, nil
}

func (k *KafkaBroker) SetEnvelopeHeader(event events.Event, envelope *events.Envelope) {
	var topicID string

	switch event.EventSource() {
	case events.EventSourceOnAdd:
		topicID = k.OnAddTopicID
	case events.EventSourceOnUpdate:
		topicID = k.OnUpdateTopicID
	case events.EventSourceOnDelete:
		topicID = k.OnDeleteTopicID
	default:
		topicID = k.GenericTopicID
	}

	envelope.AddHeader(TOPIC_ID_HEADER_KEY, topicID)
	envelope.AddHeader(EVENT_TYPE_HEADER_KEY, event.EventType().String())
}

func (k *KafkaBroker) SendMessage(envelope *events.Envelope) error {
	topicID, ok := GetTopicIDFromHeader(envelope)
	if !ok {
		return fmt.Errorf("topicID is not present on the envelope header")
	}

	messageID, err := k.publish(envelope, topicID)
	if err != nil {
		logrus.WithError(err).Errorf("error publishing message to topic %s", topicID)
		return err
	}

	logrus.WithField("broker", "kafka").Infof("message published to topicID:\"%s\" messageID:\"%s\"", topicID, messageID)

	return nil

}

// publish publishes the encoded version of the envelope as a message to the kafka topic
func (k *KafkaBroker) publish(envelope *events.Envelope, topicID string) (string, error) {
	msg, err := envelope.Encode()
	if err != nil {
		return "", fmt.Errorf("error encoding envelope: %v", err)
	}

	k.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicID,
			Partition: (int32)(kafka.PartitionAny)},
		Value: []byte(msg)}, nil)

	// Wait for delivery report
	e := <-k.Producer.Events()

	message := e.(*kafka.Message)
	if message.TopicPartition.Error != nil {
		return "", fmt.Errorf("failed to deliver message: %v\n", message.TopicPartition)
	}
	return string(message.Key), nil
}

func GetTopicIDFromHeader(envelope *events.Envelope) (string, bool) {
	if topicID, ok := envelope.Header.Headers[TOPIC_ID_HEADER_KEY]; ok {
		return topicID, true
	}

	return "", false
}

// CheckEmpty is a helper function that will check if source is empty and assign newValue if so
func CheckEmpty(source, newValue string) string {
	if source == "" {
		return newValue
	}
	return source
}

// ApplyDefaults sets default values for the Config used by the KafkaBroker
func (c *Config) ApplyDefaults() {
	c.GenericTopicID = CheckEmpty(c.GenericTopicID, DEFAULT_TOPIC_ID)
	c.OnAddTopicID = CheckEmpty(c.OnAddTopicID, DEFAULT_TOPIC_ID)
	c.OnUpdateTopicID = CheckEmpty(c.OnUpdateTopicID, DEFAULT_TOPIC_ID)
	c.OnDeleteTopicID = CheckEmpty(c.OnDeleteTopicID, DEFAULT_TOPIC_ID)
}
