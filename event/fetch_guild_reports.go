package event

import (
	"context"
	"log"

	google_pubsub "cloud.google.com/go/pubsub"
	graphql_lib "github.com/FabianHahn/graphql"
	"github.com/FabianHahn/raidlogscan/graphql"
	"github.com/FabianHahn/raidlogscan/pubsub"
	google_event "github.com/cloudevents/sdk-go/v2/event"
)

func FetchGuildReports(
	ctx context.Context,
	e google_event.Event,
	pubsubClient *google_pubsub.Client,
	graphqlClient *graphql_lib.Client,
) error {
	guildId, err := pubsub.ParseGuildReportsEvent(e)
	if err != nil {
		return err
	}

	reports, pages, err := graphql.QueryGuildReports(graphqlClient, ctx, guildId)
	if err != nil {
		return err
	}

	err = pubsub.PublishReportEvents(pubsubClient, ctx, reports)
	if err != nil {
		return err
	}

	log.Printf("Fetched %v reports in %v pages.\n", len(reports), pages)
	return nil
}
