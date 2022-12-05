package event

import (
	"context"
	"fmt"
	"log"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/cache"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/pubsub"
	google_event "github.com/cloudevents/sdk-go/v2/event"
)

func ReportAccountClaim(ctx context.Context, e google_event.Event, datastoreClient *google_datastore.Client) error {
	reportAccountClaimEvent, err := pubsub.ParseReportAccountClaimEvent(e)
	if err != nil {
		return err
	}

	reportKey := google_datastore.NameKey("report", reportAccountClaimEvent.ReportCode, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err.Error())
	}

	var report datastore.Report
	err = tx.Get(reportKey, &report)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(
			"for report account claim %v/%v datastore get report %v failed: %v",
			reportAccountClaimEvent.ClaimedAccountName,
			reportAccountClaimEvent.ClaimedPlayerId,
			reportAccountClaimEvent.ReportCode,
			err.Error())
	}

	found := false
	for i := range report.PlayerAccounts {
		if report.PlayerAccounts[i].PlayerId == reportAccountClaimEvent.ClaimedPlayerId {
			report.PlayerAccounts[i].Name = reportAccountClaimEvent.ClaimedAccountName
			found = true
			break
		}
	}
	if !found {
		report.PlayerAccounts = append(report.PlayerAccounts, datastore.ReportPlayerAccount{
			PlayerId: reportAccountClaimEvent.ClaimedPlayerId,
			Name:     reportAccountClaimEvent.ClaimedAccountName,
		})
	}

	_, err = tx.Put(reportKey, &report)
	if err != nil {
		return fmt.Errorf(
			"for report account claim %v/%v datastore write report %v failed: %v",
			reportAccountClaimEvent.ClaimedAccountName,
			reportAccountClaimEvent.ClaimedPlayerId,
			reportAccountClaimEvent.ReportCode,
			err.Error())
	}

	_, err = tx.Commit()
	if err != nil {
		return fmt.Errorf(
			"report account claim %v/%v report %v datastore transaction failed: %v",
			reportAccountClaimEvent.ClaimedAccountName,
			reportAccountClaimEvent.ClaimedPlayerId,
			reportAccountClaimEvent.ReportCode,
			err.Error())
	}

	if report.GuildId != 0 {
		err = cache.InvalidateGuildStatsCache(ctx, datastoreClient, report.GuildId)
		if err != nil {
			return fmt.Errorf("failed to invalidate guild stats cache for %v: %v", report.GuildId, err)
		}
	}

	log.Printf(
		"Updated report account claim %v/%v for report %v.\n",
		reportAccountClaimEvent.ClaimedAccountName,
		reportAccountClaimEvent.ClaimedPlayerId,
		reportAccountClaimEvent.ReportCode)
	return nil
}
