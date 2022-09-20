package pubsub

import (
	"context"
	"log"
	"os"

	google_pubsub "cloud.google.com/go/pubsub"
)

func CreatePubsubClientOrDie() *google_pubsub.Client {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")

	pubsubClient, err := google_pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	return pubsubClient
}
