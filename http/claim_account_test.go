package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/pubsub"
)

const (
	testClaimAccountName = "Jaythe"
	testClaimPlayerId    = "71133535"
)

func TestClaimAccount(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?account_name=%v&player_id=%v", testClaimAccountName, testClaimPlayerId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	playerStatsUrl := "http://example.com/playerstats"
	accountStatsUrl := "http://example.com/accountstats"
	ClaimAccount(rr, req, datastoreClient, pubsubClient, playerStatsUrl, accountStatsUrl)

	t.Log(rr.Body.String())
}
