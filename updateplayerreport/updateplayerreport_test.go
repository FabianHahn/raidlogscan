package updateplayerreport

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	testReportCode = "q1ZxbNt74DB6zFr2"
	testPlayerId   = "71133535"
)

func TestUpdatePlayerReport(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	originalFlags := log.Flags()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	message := MessagePublishedData{
		Message: PubSubMessage{
			Attributes: map[string]interface{}{
				"code":      testReportCode,
				"player_id": testPlayerId,
			},
		},
	}

	e := event.New()
	e.SetDataContentType("application/json")
	e.SetData(e.DataContentType(), message)

	err := updatePlayerReport(context.Background(), e)
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
