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

func FetchRecentCharacterReports(
	ctx context.Context,
	e google_event.Event,
	pubsubClient *google_pubsub.Client,
	graphqlClient *graphql_lib.Client,
) error {
	characterId, err := pubsub.ParseRecentCharacterReportsEvent(e)
	if err != nil {
		return err
	}

	reports, pages, err := graphql.QueryRecentCharacterReports(graphqlClient, ctx, characterId)
	if err != nil {
		return err
	}

	err = pubsub.PublishReportEvents(pubsubClient, ctx, reports)
	if err != nil {
		return err
	}

	log.Printf("Fetched %v recent character reports in %v pages.\n", len(reports), pages)
	return nil
}
