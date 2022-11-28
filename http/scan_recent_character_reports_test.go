package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/pubsub"
)

const (
	testScanRecentCharacterReportsCharacterId = "67578566"
)

func TestScanRecentCharacterReports(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?character_id=%v", testScanRecentCharacterReportsCharacterId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	ScanRecentCharacterReports(rr, req, pubsubClient)

	t.Log(rr.Body.String())
}
