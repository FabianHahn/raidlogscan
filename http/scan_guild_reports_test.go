package http

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/pubsub"
)

const (
	testScanGuildReportsGuildId = "687460"
)

func TestScanGuildReports(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?guild_id=%v", testScanGuildReportsGuildId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	guildStatsUrl := "http://example.com/guildstats"
	ScanGuildReports(rr, req, datastoreClient, pubsubClient, guildStatsUrl)

	t.Log(rr.Body.String())
}
