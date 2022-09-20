package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
)

const (
	testStatsPlayerId = "71133535"
)

func TestPlayerStats(t *testing.T) {

	req := httptest.NewRequest("GET", fmt.Sprintf("/?player_id=%v", testStatsPlayerId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	accountStatsUrl := "http://example.com/accountstats"
	claimAccountUrl := "http://example.com/claimaccount"
	PlayerStats(rr, req, datastoreClient, accountStatsUrl, claimAccountUrl)

	t.Log(rr.Body.String())
}
