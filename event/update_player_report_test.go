package event

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	testUpdateReportCode = "q1ZxbNt74DB6zFr2"
	testUpdatePlayerId   = "71133535"
)

func TestUpdatePlayerReport(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	originalFlags := log.Flags()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	message := MessagePublishedData{
		Message: PubSubMessage{
			Attributes: map[string]interface{}{
				"code":      testUpdateReportCode,
				"player_id": testUpdatePlayerId,
			},
		},
	}

	e := event.New()
	e.SetDataContentType("application/json")
	e.SetData(e.DataContentType(), message)

	datastoreClient := datastore.CreateDatastoreClientOrDie()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	err := UpdatePlayerReport(context.Background(), e, datastoreClient, pubsubClient)
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
