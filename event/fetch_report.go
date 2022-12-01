package event

import (
	"context"
	"fmt"
	"log"
	"time"

	google_datastore "cloud.google.com/go/datastore"
	google_pubsub "cloud.google.com/go/pubsub"
	graphql_lib "github.com/FabianHahn/graphql"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/graphql"
	"github.com/FabianHahn/raidlogscan/pubsub"
	google_event "github.com/cloudevents/sdk-go/v2/event"
)

func FetchReport(
	ctx context.Context,
	e google_event.Event,
	datastoreClient *google_datastore.Client,
	pubsubClient *google_pubsub.Client,
	graphqlClient *graphql_lib.Client,
) error {
	code, err := pubsub.ParseReportEvent(e)
	if err != nil {
		return err
	}

	key := google_datastore.NameKey("report", code, nil)
	var report datastore.Report
	oldVersionPlayerAccounts := []datastore.ReportPlayerAccount{}
	err = datastoreClient.Get(ctx, key, &report)
	if err != nil && err != google_datastore.ErrNoSuchEntity {
		return fmt.Errorf("datastore query for %v failed: %v", code, err.Error())
	} else if err == nil {
		if report.Version >= 2 {
			log.Printf("Report %v already processed.\n", code)
			return nil
		}
		oldVersionPlayerAccounts = report.PlayerAccounts
		log.Printf(
			"Outdated entry for report %v, replacing entry while preserving %v player accounts.\n",
			code,
			len(oldVersionPlayerAccounts))
	}

	reportQueryResult, err := graphql.QueryReport(graphqlClient, ctx, code)
	if err != nil {
		return err
	}

	report.Title = reportQueryResult.Title
	report.CreatedAt = time.Now()
	report.StartTime = reportQueryResult.StartTime
	report.EndTime = reportQueryResult.EndTime
	report.Zone = reportQueryResult.Zone
	report.GuildId = reportQueryResult.GuildId
	report.GuildName = reportQueryResult.GuildName
	report.PlayerAccounts = oldVersionPlayerAccounts
	report.Version = 2

	for _, player := range reportQueryResult.Players.Tanks {
		report.Players = append(report.Players, datastore.ReportPlayer{
			Id:     player.Guid,
			Name:   player.Name,
			Class:  player.Class,
			Server: player.Server,
			Spec:   "",
			Role:   "tank",
		})
	}
	for _, player := range reportQueryResult.Players.Dps {
		report.Players = append(report.Players, datastore.ReportPlayer{
			Id:     player.Guid,
			Name:   player.Name,
			Class:  player.Class,
			Server: player.Server,
			Spec:   player.Spec,
			Role:   "dps",
		})
	}
	for _, player := range reportQueryResult.Players.Healers {
		report.Players = append(report.Players, datastore.ReportPlayer{
			Id:     player.Guid,
			Name:   player.Name,
			Class:  player.Class,
			Server: player.Server,
			Spec:   player.Spec,
			Role:   "healer",
		})
	}

	_, err = datastoreClient.Put(ctx, key, &report)
	if err != nil {
		return fmt.Errorf("datastore write for %s failed: %v", code, err.Error())
	}

	playerIds := []int64{}
	for _, player := range report.Players {
		playerIds = append(playerIds, player.Id)
	}

	err = pubsub.PublishPlayerReportEvents(pubsubClient, ctx, code, playerIds)
	if err != nil {
		return err
	}

	log.Printf("Processed report %v.\n", code)
	return nil
}
