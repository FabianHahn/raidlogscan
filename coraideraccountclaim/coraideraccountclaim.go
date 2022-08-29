package coraideraccountclaim

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
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

type AccountCharacter struct {
	Id     int64
	Name   string
	Class  string
	Server string
}

type Account struct {
	Characters []AccountCharacter
}

var datastoreClient *datastore.Client

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
	datastoreClient, err = datastore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	functions.CloudEvent("CoraiderAccountClaim", coraiderAccountClaim)
}

// coraiderAccountClaim is an HTTP Cloud Function.
func coraiderAccountClaim(ctx context.Context, e event.Event) error {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return fmt.Errorf("failed to parse event message data: %v", err)
	}

	claimedAccountName := message.Message.Attributes["claimed_account_name"].(string)
	playerId, err := strconv.ParseInt(message.Message.Attributes["player_id"].(string), 10, 64)
	if err != nil {
		return fmt.Errorf("player ID conversion failed: %v", err.Error())
	}
	claimedPlayerId, err := strconv.ParseInt(message.Message.Attributes["claimed_player_id"].(string), 10, 64)
	if err != nil {
		return fmt.Errorf("claimed Player ID conversion failed: %v", err.Error())
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
		return fmt.Errorf("for coraider account claim %v/%v datastore get player %v failed: %v", claimedAccountName, claimedPlayerId, playerId, err.Error())
	}

	found := false
	for i := range player.CoraiderAccounts {
		if player.CoraiderAccounts[i].PlayerId == claimedPlayerId {
			player.CoraiderAccounts[i].Name = claimedAccountName
			found = true
			break
		}
	}
	if !found {
		player.CoraiderAccounts = append(player.CoraiderAccounts, PlayerCoraiderAccount{
			PlayerId: claimedPlayerId,
			Name:     claimedAccountName,
		})
	}

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		return fmt.Errorf("for coraider account claim %v/%v datastore write player %v failed: %v", claimedAccountName, claimedPlayerId, playerId, err.Error())
	}

	_, err = tx.Commit()
	if err != nil {
		return fmt.Errorf("coraider account claim %v/%v player %v datastore transaction failed: %v", claimedAccountName, claimedPlayerId, playerId, err.Error())
	}

	log.Printf("Updated coraider account claim %v/%v for player %v.\n", claimedAccountName, claimedPlayerId, playerId)
	return nil
}
