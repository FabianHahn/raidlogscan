package playerstats

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

const (
	testPlayerId = "71133535"
)

func TestPlayerStats(t *testing.T) {

	req := httptest.NewRequest("GET", fmt.Sprintf("/?player_id=%v", testPlayerId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	playerStats(rr, req)

	t.Log(rr.Body.String())
}
