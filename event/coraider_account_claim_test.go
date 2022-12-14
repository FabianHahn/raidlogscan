package event

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	testClaimedAccountName = "Jaythe"
	testClaimedPlayerId    = "38937027"
	testNotifiedPlayerId   = "11296426"
)

func TestCoraiderAccountClaim(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	originalFlags := log.Flags()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	message := MessagePublishedData{
		Message: PubSubMessage{
			Attributes: map[string]interface{}{
				"player_id":            testNotifiedPlayerId,
				"claimed_player_id":    testClaimedPlayerId,
				"claimed_account_name": testClaimedAccountName,
			},
		},
	}

	e := event.New()
	e.SetDataContentType("application/json")
	e.SetData(e.DataContentType(), message)

	datastoreClient := datastore.CreateDatastoreClientOrDie()
	err := CoraiderAccountClaim(context.Background(), e, datastoreClient)
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
