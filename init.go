package raidlogscan

import (
	"context"
	go_http "net/http"
	"os"

	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/event"
	"github.com/FabianHahn/raidlogscan/graphql"
	"github.com/FabianHahn/raidlogscan/html"
	"github.com/FabianHahn/raidlogscan/http"
	"github.com/FabianHahn/raidlogscan/oauth2"
	"github.com/FabianHahn/raidlogscan/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	google_event "github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	accountStatsUrl := os.Getenv("RAIDLOGSCAN_ACCOUNTSTATS_URL")
	claimAccountUrl := os.Getenv("RAIDLOGSCAN_CLAIMACCOUNT_URL")
	playerStatsUrl := os.Getenv("RAIDLOGSCAN_PLAYERSTATS_URL")
	guildStatsUrl := os.Getenv("RAIDLOGSCAN_GUILDSTATS_URL")
	oauth2LoginUrl := os.Getenv("RAIDLOGSCAN_OAUTH2_LOGIN_URL")
	oauth2RedirectUrl := os.Getenv("RAIDLOGSCAN_OAUTH2_REDIRECT_URL")
	scanUserReportsUrl := os.Getenv("RAIDLOGSCAN_SCAN_USER_REPORTS_URL")
	scanCharacterReportsUrl := os.Getenv("RAIDLOGSCAN_SCAN_CHARACTER_REPORTS_URL")
	scanGuildReportsUrl := os.Getenv("RAIDLOGSCAN_SCAN_GUILD_REPORTS_URL")

	oauth2UserConfig := oauth2.CreateOauth2UserConfig(oauth2RedirectUrl)
	htmlRenderer := html.CreateRendererOrDie()
	datastoreClient := datastore.CreateDatastoreClientOrDie()
	pubsubClient := pubsub.CreatePubsubClientOrDie()
	graphqlClient := graphql.CreateGraphqlClient()

	functions.CloudEvent("CoraiderAccountClaim", func(ctx context.Context, e google_event.Event) error {
		return event.CoraiderAccountClaim(ctx, e, datastoreClient)
	})
	functions.CloudEvent("ReportAccountClaim", func(ctx context.Context, e google_event.Event) error {
		return event.ReportAccountClaim(ctx, e, datastoreClient)
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
	functions.CloudEvent("FetchUserReports", func(ctx context.Context, e google_event.Event) error {
		return event.FetchUserReports(ctx, e, pubsubClient, graphqlClient)
	})
	functions.CloudEvent("FetchRecentCharacterReports", func(ctx context.Context, e google_event.Event) error {
		return event.FetchRecentCharacterReports(ctx, e, pubsubClient, graphqlClient)
	})

	functions.HTTP("AccountStats", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.AccountStats(w, r, htmlRenderer, datastoreClient, playerStatsUrl, oauth2LoginUrl)
	})
	functions.HTTP("ClaimAccount", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.ClaimAccount(w, r, datastoreClient, pubsubClient, playerStatsUrl, accountStatsUrl)
	})
	functions.HTTP("PlayerStats", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.PlayerStats(w, r, htmlRenderer, datastoreClient, accountStatsUrl, guildStatsUrl, claimAccountUrl)
	})
	functions.HTTP("GuildStats", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.GuildStats(w, r, htmlRenderer, datastoreClient, scanGuildReportsUrl, accountStatsUrl, playerStatsUrl)
	})
	functions.HTTP("Oauth2Login", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.Oauth2Login(w, r, oauth2UserConfig)
	})
	functions.HTTP("Oauth2Callback", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.Oauth2Callback(w, r, oauth2UserConfig, scanUserReportsUrl,
			scanCharacterReportsUrl)
	})
	functions.HTTP("ScanUserReports", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.ScanUserReports(w, r, pubsubClient)
	})
	functions.HTTP("ScanRecentCharacterReports", func(w go_http.ResponseWriter, r *go_http.Request) {
		http.ScanRecentCharacterReports(w, r, pubsubClient)
	})
}
