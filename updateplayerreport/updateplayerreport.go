package updateplayerreport

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	coraiderAccountClaimTopicId = "coraideraccountclaim"
	numCoraiderClaimBroadcasts  = 3
)

type PubSubMessage struct {
	Attributes map[string]interface{} `json:"attributes"`
}

type MessagePublishedData struct {
	Message PubSubMessage
}

type ReportPlayer struct {
	Id     int64
	Name   string
	Class  string
	Server string
	Spec   string
	Role   string
}

type Report struct {
	Title     string
	CreatedAt time.Time
	StartTime time.Time
	EndTime   time.Time
	Zone      string
	Players   []ReportPlayer `datastore:",noindex"`
}

type PlayerReport struct {
	Code      string
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Zone      string
	Spec      string
	Role      string
	Duplicate bool
}

type PlayerCoraider struct {
	Id     int64
	Name   string
	Class  string
	Server string
	Count  int64
}

type PlayerCoraiderAccount struct {
	Name     string
	PlayerId int64
}

type Player struct {
	Name             string
	Class            string
	Server           string
	Account          string
	Reports          []PlayerReport          `datastore:",noindex"`
	Coraiders        []PlayerCoraider        `datastore:",noindex"`
	CoraiderAccounts []PlayerCoraiderAccount `datastore:",noindex"`
	Version          int64
}

var datastoreClient *datastore.Client
var pubsubClient *pubsub.Client

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
	datastoreClient, err = datastore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	pubsubClient, err = pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	functions.CloudEvent("UpdatePlayerReport", updatePlayerReport)
}

// fetchReport is an HTTP Cloud Function.
func updatePlayerReport(ctx context.Context, e event.Event) error {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return fmt.Errorf("failed to parse event message data: %v", err)
	}

	code := message.Message.Attributes["code"].(string)
	playerId, err := strconv.ParseInt(message.Message.Attributes["player_id"].(string), 10, 64)
	if err != nil {
		return fmt.Errorf("Player ID conversion failed: %v", err.Error())
	}

	reportKey := datastore.NameKey("report", code, nil)
	var report Report
	err = datastoreClient.Get(ctx, reportKey, &report)
	if err != nil {
		return fmt.Errorf("datastore report query %v failed: %v", code, err.Error())
	}

	if !report.EndTime.After(report.StartTime) {
		log.Printf("Got empty report %v, not updating player %v.\n", code, playerId)
		return nil
	}

	var thisReportPlayer *ReportPlayer
	for _, player := range report.Players {
		if player.Id == playerId {
			thisReportPlayer = &player
			break
		}
	}
	if thisReportPlayer == nil {
		return fmt.Errorf("Player %v not found in report: %+v", playerId, report)
	}

	playerKey := datastore.IDKey("player", playerId, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err.Error())
	}

	var player Player
	err = tx.Get(playerKey, &player)
	if err == datastore.ErrNoSuchEntity {
		player.Name = thisReportPlayer.Name
		player.Class = thisReportPlayer.Class
		player.Server = thisReportPlayer.Server
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf("for update report %v datastore get player %v failed: %v", code, playerId, err.Error())
	}

	if player.Version < 2 {
		log.Printf("Outdated entry for player %v, replacing entry.\n", playerId)
		player.Reports = []PlayerReport{}
		player.Coraiders = []PlayerCoraider{}
		player.Version = 2
	}

	for _, playerReport := range player.Reports {
		if playerReport.Code == code {
			tx.Rollback()
			log.Printf("Report %v already reported for player %v.\n", code, playerId)
			return nil // no error
		}
	}

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

	player.Reports = append(player.Reports, PlayerReport{
		Code:      code,
		Title:     report.Title,
		StartTime: report.StartTime,
		EndTime:   report.EndTime,
		Zone:      report.Zone,
		Spec:      thisReportPlayer.Spec,
		Role:      thisReportPlayer.Role,
		Duplicate: duplicate,
	})
	sort.SliceStable(player.Reports, func(i int, j int) bool {
		return player.Reports[i].StartTime.After(player.Reports[j].StartTime)
	})

	newCoraiderIds := []int64{}
	if !duplicate {
		coraiders := map[int64]*PlayerCoraider{}
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
				coraiders[reportPlayer.Id] = &PlayerCoraider{
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

		player.Coraiders = []PlayerCoraider{}
		for _, coraider := range coraiders {
			player.Coraiders = append(player.Coraiders, *coraider)
		}
	}

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update player when updating report %v for player %v: %v", code, playerId, err.Error())
	}

	_, err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction updating report %v for player %v: %v", code, playerId, err.Error())
	}

	if player.Account != "" && len(newCoraiderIds) > 0 {
		var waitGroup sync.WaitGroup
		var totalErrors uint64
		coraiderAccountClaimTopic := pubsubClient.Topic(coraiderAccountClaimTopicId)
		for _, newCoraiderId := range newCoraiderIds {
			result := coraiderAccountClaimTopic.Publish(ctx, &pubsub.Message{
				Attributes: map[string]string{
					"player_id":            strconv.FormatInt(newCoraiderId, 10),
					"claimed_player_id":    strconv.FormatInt(playerId, 10),
					"claimed_account_name": player.Account,
				},
			})

			waitGroup.Add(1)
			go func(res *pubsub.PublishResult) {
				defer waitGroup.Done()
				// The Get method blocks until a server-generated ID or
				// an error is returned for the published message.
				_, err := res.Get(ctx)
				if err != nil {
					// Error handling code can be added here.
					log.Printf("Failed to publish: %v\n", err)
					atomic.AddUint64(&totalErrors, 1)
					return
				}
			}(result)
		}
		waitGroup.Wait()

		if totalErrors > 0 {
			return fmt.Errorf("failed to update report %v for player %v: %d pubsub writes failed", code, playerId, totalErrors)
		}

		log.Printf("Processed report %v for player %v and broadcast account to %v new coraiders.\n", code, playerId, len(newCoraiderIds))
		return nil
	}

	log.Printf("Processed report %v for player %v.\n", code, playerId)
	return nil
}
