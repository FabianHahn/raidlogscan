package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
)

const (
	testStatsAccountName = "Jaythe"
)

func TestAccountStats(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?account_name=%v", testStatsAccountName), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	playerStatsUrl := "http://example.com/playerstats"
	AccountStats(rr, req, datastoreClient, playerStatsUrl)

	t.Log(rr.Body.String())
}