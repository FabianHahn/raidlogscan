package raidlogscan

import (
	"context"
	go_http "net/http"
	"os"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/event"
	"github.com/FabianHahn/raidlogscan/graphql"
	"github.com/FabianHahn/raidlogscan/http"
	"github.com/FabianHahn/raidlogscan/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	google_event "github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	graphqlClient := graphql.CreateGraphqlClientOrDie()

	accountStatsUrl := os.Getenv("RAIDLOGSCAN_ACCOUNTSTATS_URL")
	claimAccountUrl := os.Getenv("RAIDLOGSCAN_CLAIMACCOUNT_URL")
	playerStatsUrl := os.Getenv("RAIDLOGSCAN_PLAYERSTATS_URL")

	functions.CloudEvent("CoraiderAccountClaim", func(ctx context.Context, e google_event.Event) error {
		return event.CoraiderAccountClaim(ctx, e, datastoreClient)
	})
	functions.CloudEvent("FetchGuildReports", func(ctx context.Context, e google_event.Event) error {
		return event.FetchGuildReports(ctx, e, pubsubClient, graphqlClient)
	})
	functions.CloudEvent("FetchReport", func(ctx context.Context, e google_event.Event) error {
		return event.FetchReport(ctx, e, datastoreClient, pubsubClient, graphqlClient)
	})
	functions.CloudEvent("UpdatePlayerReport", func(ctx context.Context, e google_event.Event) error {
		return event.UpdatePlayerReport(ctx, e, datastoreClient, pubsubClient)
	})

	functions.HTTP("AccountStats", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.AccountStats(w, r, datastoreClient, playerStatsUrl)
	})
	functions.HTTP("ClaimAccount", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.ClaimAccount(w, r, datastoreClient, pubsubClient, playerStatsUrl, accountStatsUrl)
	})
	functions.HTTP("PlayerStats", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.PlayerStats(w, r, datastoreClient, accountStatsUrl, claimAccountUrl)
	})
}