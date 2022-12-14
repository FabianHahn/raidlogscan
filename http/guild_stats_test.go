package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/html"
)

const (
	testStatsGuildId = "687460"
)

func TestGuildStats(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?guild_id=%v", testStatsGuildId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	htmlRenderer := html.CreateRendererOrDie()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	scanGuildReportsUrl := "http://example.com/scanguildreports"
	accountStatsUrl := "http://example.com/accountstats"
	playerStatsUrl := "http://example.com/playerstats"
	oauth2LoginUrl := "http://example.com/oauth2login"
	GuildStats(
		rr,
		req,
		htmlRenderer,
		datastoreClient,
		scanGuildReportsUrl,
		accountStatsUrl,
		playerStatsUrl,
		oauth2LoginUrl)

	t.Log(rr.Body.String())
}
