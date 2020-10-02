package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func Test_PubSubBroker_BuildEnvelope_GameServerEvent(t *testing.T) {
	projectID := "calm-weather-345673"

	testCases := []struct {
		desc    string
		topicID string
		event   func(message events.Message) events.Event
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
						PROJECTID_HEADER_KEY:  projectID,
						EVENT_TYPE_HEADER_KEY: events.GameServerEventAdded.String(),
						TOPIC_ID_HEADER_KEY:   "gameserver.events",
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
						PROJECTID_HEADER_KEY:  projectID,
						EVENT_TYPE_HEADER_KEY: events.GameServerEventUpdated.String(),
						TOPIC_ID_HEADER_KEY:   "gameserver.events",
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
						PROJECTID_HEADER_KEY:  projectID,
						EVENT_TYPE_HEADER_KEY: events.GameServerEventDeleted.String(),
						TOPIC_ID_HEADER_KEY:   "gameserver.events",
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
			require.Equal(t, reflect.DeepEqual(got, tc.expect), true)
		})
	}
}

func Test_PubSubBroker_SendMessage(t *testing.T) {
	// TODO: Mock Pub/Sub client
	t.SkipNow()

	t.Parallel()

	t.Run("it should send a message to a topic that exists", func(t *testing.T) {
		projectID := "calm-weather-345673"
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
					PROJECTID_HEADER_KEY:  projectID,
					EVENT_TYPE_HEADER_KEY: events.GameServerEventAdded.String(),
					TOPIC_ID_HEADER_KEY:   topicID,
				},
			},
			Message: "fakeBody",
		}

		err = broker.SendMessage(envelope)
		require.Nil(t, err)
	})

	t.Run("it should not send a message to a topic that does not exist", func(t *testing.T) {
		projectID := "calm-weather-345673"
		topicID := "none"

		_ = setup(t, projectID, topicID)

		config := &Config{
			ProjectID: projectID,
		}

		broker, err := NewPubSubBroker(config)
		require.Nil(t, err)

		envelope := &events.Envelope{
			Header: &events.Header{
				Headers: map[string]string{
					PROJECTID_HEADER_KEY:  projectID,
					EVENT_TYPE_HEADER_KEY: events.GameServerEventAdded.String(),
					TOPIC_ID_HEADER_KEY:   topicID,
				},
			},
			Message: "fakeBody",
		}

		err = broker.SendMessage(envelope)
		require.NotNil(t, err)
	})
}

func Test_GetTopicIDFromHeader(t *testing.T) {
	type want struct {
		TopicID string
		Ok      bool
	}

	testCases := []struct {
		desc   string
		header map[string]string
		want   want
	}{
		{
			desc: "it should return topicID from single header",
			header: map[string]string{
				TOPIC_ID_HEADER_KEY: "gameserver.events",
			},
			want: want{
				TopicID: "gameserver.events",
				Ok:      true,
			},
		},
		{
			desc: "it should return topicID from multiples header",
			header: map[string]string{
				TOPIC_ID_HEADER_KEY:   "gameserver.events",
				EVENT_TYPE_HEADER_KEY: "Added",
			},
			want: want{
				TopicID: "gameserver.events",
				Ok:      true,
			},
		},
		{
			desc:   "it should not return topicID from a empty header",
			header: map[string]string{},
			want: want{
				TopicID: "",
				Ok:      false,
			},
		},
		{
			desc: "it should not return topicID from non empty header and missing topic_id",
			header: map[string]string{
				"header1": "1",
				"header2": "2",
			},
			want: want{
				TopicID: "",
				Ok:      false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			envelope := &events.Envelope{
				Header: &events.Header{
					Headers: tc.header,
				},
				Message: "fakeBody",
			}

			got, err := GetTopicIDFromHeader(envelope)
			assert.Equal(t, got, tc.want.TopicID)
			assert.Equal(t, err, tc.want.Ok)
		})
	}
}

func setup(t *testing.T, projectID, topicID string) *pubsub.Client {
	ctx := context.Background()

	var err error
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Cleanup resources from the previous tests.
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

	return client
}
