/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

// Create topics using the local Pub/Sub emulator.
// Make sure you have the environment variable PUBSUB_EMULATOR_HOST set
// using the information from the emulator output
// https://cloud.google.com/pubsub/docs/emulator

// Usage: go run create_topics.go
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
