package claimaccount

import (
	"context"
	"fmt"
	"log"
	"os"
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

	functions.CloudEvent("ClaimAccount", claimAccount)
}

// claimAccount is an HTTP Cloud Function.
func claimAccount(ctx context.Context, e event.Event) error {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return fmt.Errorf("failed to parse event message data: %v", err)
	}

	accountName := message.Message.Attributes["account_name"].(string)
	playerId, err := strconv.ParseInt(message.Message.Attributes["player_id"].(string), 10, 64)
	if err != nil {
		return fmt.Errorf("player ID conversion failed: %v", err.Error())
	}

	playerKey := datastore.IDKey("player", playerId, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err.Error())
	}

	var player Player
	err = tx.Get(playerKey, &player)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("for claim account %v datastore get player %v failed: %v", accountName, playerId, err.Error())
	}

	player.Account = accountName

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("datastore write claim account %v player %v failed: %v", accountName, playerId, err.Error())
	}

	_, err = tx.Commit()
	if err != nil {
		return fmt.Errorf("datastore write claim account %v player %v failed: %v", accountName, playerId, err.Error())
	}

	var waitGroup sync.WaitGroup
	var totalErrors uint64
	coraiderAccountClaimTopic := pubsubClient.Topic(coraiderAccountClaimTopicId)
	for _, coraider := range player.Coraiders {
		result := coraiderAccountClaimTopic.Publish(ctx, &pubsub.Message{
			Attributes: map[string]string{
				"player_id":            strconv.FormatInt(coraider.Id, 10),
				"claimed_player_id":    strconv.FormatInt(playerId, 10),
				"claimed_account_name": accountName,
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
		return fmt.Errorf("failed to claim account %v player %v: %d pubsub writes failed", accountName, playerId, totalErrors)
	}

	log.Printf("Claimed account name %v for player %v.\n", accountName, playerId)
	return nil
}
