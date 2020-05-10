package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/Octops/gameserver-events-broadcaster/pkg/events"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestPubSubBroker_BuildEnvelope_GameServerEvent(t *testing.T) {
	projectID := "nice-storm-235718"

	testCases := []struct {
		desc    string
		topicID string
		event   func(message events.Message) *events.GameServerEvent
		message events.Message
		expect  *events.Envelope
	}{
		{
			desc:    "it should build envelope for GameServerAdded Event",
			topicID: "gameserver.events.added",
			event:   events.GameServerAdded,
			message: &events.EventMessage{Body: "fakeBody"},
			expect: &events.Envelope{
				Header: &events.Header{
					Headers: map[string]string{
						"event_type": events.GameServerEventAdded,
						"project_id": projectID,
						"topic_id":   "gameserver.events",
					},
				},
				Message: "fakeBody",
			},
		},
		{
			desc:    "it should build envelope for GameServerUpdated Event",
			topicID: "gameserver.events.updated",
			event:   events.GameServerUpdated,
			message: &events.EventMessage{Body: "fakeBody"},
			expect: &events.Envelope{
				Header: &events.Header{
					Headers: map[string]string{
						"event_type": events.GameServerEventUpdated,
						"project_id": projectID,
						"topic_id":   "gameserver.events",
					},
				},
				Message: "fakeBody",
			},
		},
		{
			desc:    "it should build envelope for GameServerDeleted Event",
			topicID: "gameserver.events.deleted",
			event:   events.GameServerDeleted,
			message: &events.EventMessage{Body: "fakeBody"},
			expect: &events.Envelope{
				Header: &events.Header{
					Headers: map[string]string{
						"event_type": events.GameServerEventDeleted,
						"project_id": projectID,
						"topic_id":   "gameserver.events",
					},
				},
				Message: "fakeBody",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config := &Config{
				ProjectID: projectID,
			}

			broker, err := NewPubSubBroker(config)
			require.Nil(t, err)

			message := tc.message
			event := tc.event(message)

			got, err := broker.BuildEnvelope(event)
			require.Nil(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestPubSubBroker_SendMessage(t *testing.T) {
	t.Run("it should send a message to a topic that exists", func(t *testing.T) {
		projectID := "nice-storm-235718"
		topicID := "gameserver.events"

		client := setup(t, projectID, topicID)
		_, err := client.CreateTopic(context.Background(), topicID)
		if err != nil {
			t.Error(err)
		}

		config := &Config{
			ProjectID: projectID,
		}

		broker, err := NewPubSubBroker(config)
		require.Nil(t, err)

		envelope := &events.Envelope{
			Header: &events.Header{
				Headers: map[string]string{
					"event_type": events.GameServerEventAdded,
					"project_id": projectID,
					"topic_id":   topicID,
				},
			},
			Message: "fakeBody",
		}

		err = broker.SendMessage(envelope)
		require.Nil(t, err)
	})
}

var once sync.Once

func setup(t *testing.T, projectID, topicID string) *pubsub.Client {
	ctx := context.Background()

	var err error
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Cleanup resources from the previous tests.
	once.Do(func() {
		topic := client.Topic(topicID)
		ok, err := topic.Exists(ctx)
		if err != nil {
			t.Fatalf("failed to check if topic exists: %v", err)
		}
		if ok {
			if err := topic.Delete(ctx); err != nil {
				t.Fatalf("failed to cleanup the topic (%q): %v", topicID, err)
			}
		}
	})

	return client
}
