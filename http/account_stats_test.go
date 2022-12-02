package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/html"
)

const (
	testStatsAccountName = "Jaythe"
)

func TestAccountStats(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?account_name=%v", testStatsAccountName), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	htmlRenderer := html.CreateRendererOrDie()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	playerStatsUrl := "http://example.com/playerstats"
	guildStatsUrl := "http://example.com/guildstats"
	oauth2LoginUrl := "http://example.com/oauth2login"
	AccountStats(rr, req, htmlRenderer, datastoreClient, playerStatsUrl, guildStatsUrl, oauth2LoginUrl)

	t.Log(rr.Body.String())
}
