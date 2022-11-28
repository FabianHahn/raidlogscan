package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/pubsub"
)

const (
	testScanUserReportsUserId = "1258790"
)

func TestScanUserReports(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?user_id=%v", testScanUserReportsUserId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	ScanUserReports(rr, req, pubsubClient)

	t.Log(rr.Body.String())
}
