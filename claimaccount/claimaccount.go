package claimaccount

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
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
var playerstatsUrl string
var accountstatsUrl string

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

	playerstatsUrl = os.Getenv("RAIDLOGCOUNT_PLAYERSTATS_URL")
	accountstatsUrl = os.Getenv("RAIDLOGCOUNT_ACCOUNTSTATS_URL")

	functions.HTTP("ClaimAccount", claimAccount)
}

// claimAccount is an HTTP Cloud Function.
func claimAccount(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	accountName := r.URL.Query().Get("account_name")
	playerId, err := strconv.ParseInt(r.URL.Query().Get("player_id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "player ID conversion failed: %v", err.Error())
		return
	}

	query := datastore.NewQuery("player").FilterField("Name", "=", accountName)
	count, err := datastoreClient.Count(ctx, query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "player by name %v lookup failed: %v", accountName, err.Error())
		return
	}
	if count == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "cannot claim account name %v that doesn't correspond to a known character name", accountName)
		return
	}

	playerKey := datastore.IDKey("player", playerId, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to create transaction: %v", err.Error())
		return
	}

	var player Player
	err = tx.Get(playerKey, &player)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "for claim account %v datastore get player %v failed: %v", accountName, playerId, err.Error())
		return
	}

	player.Account = accountName

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "datastore write claim account %v player %v failed: %v", accountName, playerId, err.Error())
		return
	}

	_, err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "datastore write claim account %v player %v failed: %v", accountName, playerId, err.Error())
		return
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
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "failed to publish: %v\n", err)
				atomic.AddUint64(&totalErrors, 1)
				return
			}
		}(result)
	}
	waitGroup.Wait()

	if totalErrors > 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to claim account %v player %v: %d pubsub writes failed", accountName, playerId, totalErrors)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "Successfully assigned character <a href=\"%v?player_id=%v\">%v-%v (%v)</a> to player #<a href=\"%v?account_name=%v\">%v</a>.<br>\n",
		playerstatsUrl,
		playerId,
		player.Name,
		player.Server,
		player.Class,
		accountstatsUrl,
		accountName,
		accountName,
	)
}
