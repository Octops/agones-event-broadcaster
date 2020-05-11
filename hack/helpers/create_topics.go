package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"log"
	"os"
)

var (
	topics = map[string]string{
		"OnAdd":    "gameserver.events.added",
		"OnUpdate": "gameserver.events.updated",
		"OnDelete": "gameserver.events.deleted",
	}
)

func main() {
	var err error

	projectID := os.Getenv("PUBSUB_PROJECT_ID")
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	for _, topicID := range topics {
		topic := client.Topic(topicID)
		ok, err := topic.Exists(ctx)
		if err != nil {
			log.Fatalf("failed to check if topic exists: %v", err)
		}
		if ok {
			log.Printf("Deleting topic %s\n", topicID)
			if err := topic.Delete(ctx); err != nil {
				log.Fatalf("failed to cleanup the topic (%q): %v", topicID, err)
			}
		}

		log.Printf("Creating topic %s\n", topicID)
		_, err = client.CreateTopic(context.Background(), topicID)
		if err != nil {
			log.Fatalf("could not create topic %s: %v", topicID, err)
		}
	}
}
