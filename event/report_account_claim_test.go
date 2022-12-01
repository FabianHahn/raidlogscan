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
	testReportAccountClaimReportCode = "c7wnfkhaFWTzv812"
	testReportClaimedAccountName     = "Jaythe"
	testReportClaimedPlayerId        = "71188939"
)

func TestReportAccountClaim(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	originalFlags := log.Flags()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	message := MessagePublishedData{
		Message: PubSubMessage{
			Attributes: map[string]interface{}{
				"report_code":          testReportAccountClaimReportCode,
				"claimed_player_id":    testReportClaimedPlayerId,
				"claimed_account_name": testReportClaimedAccountName,
			},
		},
	}

	e := event.New()
	e.SetDataContentType("application/json")
	e.SetData(e.DataContentType(), message)

	datastoreClient := datastore.CreateDatastoreClientOrDie()
	err := ReportAccountClaim(context.Background(), e, datastoreClient)
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
