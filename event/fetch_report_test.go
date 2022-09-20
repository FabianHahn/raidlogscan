package event

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/graphql"
	"github.com/FabianHahn/raidlogscan/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	testFetchReportCode = "3BybK6aLdgtMJrGN"
)

func TestFetchReport(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	originalFlags := log.Flags()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	message := pubsub.MessagePublishedData{
		Message: google_pubsub.Message{
			Attributes: map[string]string{
				"code": testFetchReportCode,
			},
		},
	}

	e := event.New()
	e.SetDataContentType("application/json")
	e.SetData(e.DataContentType(), message)

	datastoreClient := datastore.CreateDatastoreClientOrDie()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	graphqlClient := graphql.CreateGraphqlClientOrDie()
	err := FetchReport(context.Background(), e, datastoreClient, pubsubClient, graphqlClient)
	if err != nil {
		t.Fatal(err)
	}

	w.Close()
	log.SetOutput(os.Stderr)
	log.SetFlags(originalFlags)

	out, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	t.Log(string(out))
}
