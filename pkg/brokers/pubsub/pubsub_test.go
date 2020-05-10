package pubsub

import (
	"github.com/Octops/gameserver-events-broadcaster/pkg/events"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPubSubBroker_BuildEnvelope_GameServerEvent(t *testing.T) {
	projectID := "fakeProjectID"

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

			broker := NewPubSubBroker(config)

			message := tc.message
			event := tc.event(message)
			got, err := broker.BuildEnvelope(event)

			require.Nil(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}
