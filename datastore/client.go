package datastore

import (
	"context"
	"log"
	"os"

	google_datastore "cloud.google.com/go/datastore"
)

func CreateDatastoreClientOrDie() *google_datastore.Client {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")

	datastoreClient, err := google_datastore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	return datastoreClient
}
