package updateplayerreport

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

const (
	testReportCode = "2jqkMzXL37JQaVgA"
	testPlayerId   = "71133535"
)

func TestUpdatePlayerReport(t *testing.T) {

	req := httptest.NewRequest("GET", fmt.Sprintf("/?code=%v&player_id=%v", testReportCode, testPlayerId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	updatePlayerReport(rr, req)

	t.Log(rr.Body.String())
}
