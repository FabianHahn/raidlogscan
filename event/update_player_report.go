package event

import (
	"context"
	"fmt"
	"log"
	"sort"

	google_datastore "cloud.google.com/go/datastore"
	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/pubsub"
	google_event "github.com/cloudevents/sdk-go/v2/event"
)

const (
	numCoraiderClaimBroadcasts = 3
)

type PubSubMessage struct {
	Attributes map[string]interface{} `json:"attributes"`
}

type MessagePublishedData struct {
	Message PubSubMessage
}

func UpdatePlayerReport(
	ctx context.Context,
	e google_event.Event,
	datastoreClient *google_datastore.Client,
	pubsubClient *google_pubsub.Client,
) error {
	playerReportEvent, err := pubsub.ParsePlayerReportEvent(e)
	if err != nil {
		return err
	}

	reportKey := google_datastore.NameKey("report", playerReportEvent.Code, nil)
	var report datastore.Report
	err = datastoreClient.Get(ctx, reportKey, &report)
	if err != nil {
		return fmt.Errorf("datastore report query %v failed: %v", playerReportEvent.Code, err.Error())
	}

	if !report.EndTime.After(report.StartTime) {
		log.Printf(
			"Got empty report %v, not updating player %v.\n",
			playerReportEvent.Code,
			playerReportEvent.PlayerId)
		return nil
	}

	var thisReportPlayer *datastore.ReportPlayer
	for _, player := range report.Players {
		if player.Id == playerReportEvent.PlayerId {
			thisReportPlayer = &player
			break
		}
	}
	if thisReportPlayer == nil {
		return fmt.Errorf("player %v not found in report: %+v", playerReportEvent.PlayerId, report)
	}

	playerKey := google_datastore.IDKey("player", playerReportEvent.PlayerId, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err.Error())
	}

	var player datastore.Player
	err = tx.Get(playerKey, &player)
	if err == google_datastore.ErrNoSuchEntity {
		player.Name = thisReportPlayer.Name
		player.Class = thisReportPlayer.Class
		player.Server = thisReportPlayer.Server
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf(
			"for update report %v datastore get player %v failed: %v",
			playerReportEvent.Code,
			playerReportEvent.PlayerId,
			err.Error())
	}

	if player.Version < 2 {
		log.Printf("Outdated entry for player %v, replacing entry.\n", playerReportEvent.PlayerId)
		player.Reports = []datastore.PlayerReport{}
		player.Coraiders = []datastore.PlayerCoraider{}
		player.Version = 2
	}

	onlyUpdateReports := false
	for playerIndex, playerReport := range player.Reports {
		if playerReport.Code == playerReportEvent.Code {
			if playerReport.GuildId == report.GuildId && playerReport.GuildName == report.GuildName {
				tx.Rollback()
				log.Printf("Report %v already reported for player %v.\n", playerReportEvent.Code, playerReportEvent.PlayerId)
				return nil // no error
			}

			// This report's version got updated and we need to fill in the guild ID and name.
			player.Reports[playerIndex].GuildId = report.GuildId
			player.Reports[playerIndex].GuildName = report.GuildName
			onlyUpdateReports = true
		}
	}

	newCoraiderIds := []int64{}
	if !onlyUpdateReports {
		duplicate := false
		firstEarlierStarting := sort.Search(len(player.Reports), func(i int) bool {
			return player.Reports[i].StartTime.Before(report.StartTime)
		})
		if firstEarlierStarting < len(player.Reports) && report.StartTime.Before(player.Reports[firstEarlierStarting].EndTime) {
			duplicate = true
		}
		if firstEarlierStarting > 0 && report.EndTime.After(player.Reports[firstEarlierStarting-1].StartTime) {
			duplicate = true
		}

		player.Reports = append(player.Reports, datastore.PlayerReport{
			Code:      playerReportEvent.Code,
			Title:     report.Title,
			StartTime: report.StartTime,
			EndTime:   report.EndTime,
			Zone:      report.Zone,
			GuildId:   report.GuildId,
			GuildName: report.GuildName,
			Spec:      thisReportPlayer.Spec,
			Role:      thisReportPlayer.Role,
			Duplicate: duplicate,
		})
		sort.SliceStable(player.Reports, func(i int, j int) bool {
			return player.Reports[i].StartTime.After(player.Reports[j].StartTime)
		})

		if !duplicate {
			coraiders := map[int64]*datastore.PlayerCoraider{}
			for id := range player.Coraiders {
				coraider := &player.Coraiders[id]
				coraiders[coraider.Id] = coraider
			}

			currentCoraiders := map[int64]struct{}{}
			for _, reportPlayer := range report.Players {
				if _, alreadyCounted := currentCoraiders[reportPlayer.Id]; alreadyCounted {
					continue
				}

				if coraider, ok := coraiders[reportPlayer.Id]; ok {
					coraider.Count++

					if coraider.Count <= numCoraiderClaimBroadcasts {
						newCoraiderIds = append(newCoraiderIds, reportPlayer.Id)
					}
				} else {
					coraiders[reportPlayer.Id] = &datastore.PlayerCoraider{
						Id:     reportPlayer.Id,
						Name:   reportPlayer.Name,
						Class:  reportPlayer.Class,
						Server: reportPlayer.Server,
						Count:  1,
					}
					newCoraiderIds = append(newCoraiderIds, reportPlayer.Id)
				}

				currentCoraiders[reportPlayer.Id] = struct{}{}
			}

			player.Coraiders = []datastore.PlayerCoraider{}
			for _, coraider := range coraiders {
				player.Coraiders = append(player.Coraiders, *coraider)
			}
		}
	}

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(
			"failed to update player when updating report %v for player %v: %v",
			playerReportEvent.Code,
			playerReportEvent.PlayerId,
			err.Error())
	}

	_, err = tx.Commit()
	if err != nil {
		return fmt.Errorf(
			"failed to commit transaction updating report %v for player %v: %v",
			playerReportEvent.Code,
			playerReportEvent.PlayerId,
			err.Error())
	}

	if player.Account != "" && len(newCoraiderIds) > 0 {
		err = pubsub.PublishCoraiderAccountClaimEvents(
			pubsubClient,
			ctx,
			playerReportEvent.PlayerId,
			player.Account,
			newCoraiderIds)
		if err != nil {
			return fmt.Errorf(
				"failed to update report %v for player %v: %v",
				playerReportEvent.Code,
				playerReportEvent.PlayerId,
				err.Error())
		}
	}

	if player.Account != "" && report.GuildId != 0 {
		err = pubsub.PublishReportAccountClaimEvent(
			pubsubClient,
			ctx,
			playerReportEvent.PlayerId,
			player.Account,
			[]string{playerReportEvent.Code},
		)
		if err != nil {
			return fmt.Errorf(
				"failed to update report %v for player %v: %v",
				playerReportEvent.Code,
				playerReportEvent.PlayerId,
				err.Error())
		}
	}

	if onlyUpdateReports {
		log.Printf("Updated report %v for player %v and broadcast account to %v new coraiders.\n",
			playerReportEvent.Code,
			playerReportEvent.PlayerId,
			len(newCoraiderIds))
	} else {
		log.Printf("Processed report %v for player %v and broadcast account to %v new coraiders.\n",
			playerReportEvent.Code,
			playerReportEvent.PlayerId,
			len(newCoraiderIds))
	}
	return nil
}
