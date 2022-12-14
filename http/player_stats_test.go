package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/html"
)

const (
	testStatsPlayerId = "71133535"
)

func TestPlayerStats(t *testing.T) {

	req := httptest.NewRequest("GET", fmt.Sprintf("/?player_id=%v", testStatsPlayerId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	htmlRenderer := html.CreateRendererOrDie()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	accountStatsUrl := "http://example.com/accountstats"
	guildStatsUrl := "http://example.com/guildstats"
	claimAccountUrl := "http://example.com/claimaccount"
	oauth2LoginUrl := "http://example.com/oauth2login"
	PlayerStats(
		rr,
		req,
		htmlRenderer,
		datastoreClient,
		accountStatsUrl,
		guildStatsUrl,
		claimAccountUrl,
		oauth2LoginUrl,
	)

	t.Log(rr.Body.String())
}
