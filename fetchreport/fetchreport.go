package fetchreport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"github.com/FabianHahn/graphql"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	graphqlApiUrl       = "https://www.warcraftlogs.com/api/v2/client"
	oauthApiUrl         = "https://www.warcraftlogs.com/oauth"
	playerreportTopicId = "playerreport"
)

type MessagePublishedData struct {
	Message pubsub.Message
}

type Player struct {
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
	Players   []Player `datastore:",noindex"`
}

type PlayerDetailsResponse struct {
	Data struct {
		PlayerDetails struct {
			Tanks []struct {
				Name   string `json:"name"`
				Guid   int64  `json:"guid"`
				Class  string `json:"type"`
				Server string `json:"server"`
			} `json:"tanks"`
			Dps []struct {
				Name   string `json:"name"`
				Guid   int64  `json:"guid"`
				Class  string `json:"type"`
				Server string `json:"server"`
				Spec   string `json:"icon"`
			} `json:"dps"`
			Healers []struct {
				Name   string `json:"name"`
				Guid   int64  `json:"guid"`
				Class  string `json:"type"`
				Server string `json:"server"`
				Spec   string `json:"icon"`
			} `json:"healers"`
		} `json:"playerDetails"`
	} `json:"data"`
}

type ReportQuery struct {
	ReportData struct {
		Report struct {
			Title     graphql.String
			StartTime graphql.Float
			EndTime   graphql.Float
			Zone      struct {
				Name graphql.String
			}
			PlayerDetails json.RawMessage `graphql:"playerDetails(endTime: 999999999999)"`
		} `graphql:"report(code: $code)"`
	}
}

var datastoreClient *datastore.Client
var pubsubClient *pubsub.Client
var graphqlClient *graphql.Client

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

	config := clientcredentials.Config{
		ClientID:     os.Getenv("WARCRAFTLOGS_CLIENT_ID"),
		ClientSecret: os.Getenv("WARCRAFTLOGS_CLIENT_SECRET"),
		Scopes:       []string{},
		TokenURL:     oauthApiUrl + "/token",
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	oauthClient := config.Client(context.Background())
	graphqlClient = graphql.NewClient(graphqlApiUrl, oauthClient)

	functions.CloudEvent("FetchReport", fetchReport)
}

func convertFloatTime(floatTime float64) time.Time {
	integral, fractional := math.Modf(floatTime / 1000)
	return time.Unix(int64(integral), int64(fractional*1e9))
}

// fetchReport is an HTTP Cloud Function.
func fetchReport(ctx context.Context, e event.Event) error {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return fmt.Errorf("failed to parse event message data: %v", err)
	}
	code := message.Message.Attributes["code"]

	key := datastore.NameKey("report", code, nil)
	var report Report
	err := datastoreClient.Get(ctx, key, &report)
	if err == nil {
		log.Printf("Report %v already processed.\n", code)
		return nil
	} else if err != datastore.ErrNoSuchEntity {
		return fmt.Errorf("datastore query for %v failed: %v", code, err.Error())
	} else {
		var query ReportQuery
		variables := map[string]interface{}{
			"code": graphql.String(code),
		}
		err = graphqlClient.Query(ctx, &query, variables)
		if err != nil {
			return fmt.Errorf("GraphQL query for %v failed: %v", code, err.Error())
		}

		report.Title = string(query.ReportData.Report.Title)
		report.CreatedAt = time.Now()
		report.StartTime = convertFloatTime(float64(query.ReportData.Report.StartTime))
		report.EndTime = convertFloatTime(float64(query.ReportData.Report.EndTime))
		report.Zone = string(query.ReportData.Report.Zone.Name)

		var playerDetailsResponse PlayerDetailsResponse
		err = json.Unmarshal(query.ReportData.Report.PlayerDetails, &playerDetailsResponse)
		if err == nil {
			for _, player := range playerDetailsResponse.Data.PlayerDetails.Tanks {
				report.Players = append(report.Players, Player{
					Id:     player.Guid,
					Name:   player.Name,
					Class:  player.Class,
					Server: player.Server,
					Spec:   "",
					Role:   "tank",
				})
			}
			for _, player := range playerDetailsResponse.Data.PlayerDetails.Dps {
				report.Players = append(report.Players, Player{
					Id:     player.Guid,
					Name:   player.Name,
					Class:  player.Class,
					Server: player.Server,
					Spec:   player.Spec,
					Role:   "dps",
				})
			}
			for _, player := range playerDetailsResponse.Data.PlayerDetails.Healers {
				report.Players = append(report.Players, Player{
					Id:     player.Guid,
					Name:   player.Name,
					Class:  player.Class,
					Server: player.Server,
					Spec:   player.Spec,
					Role:   "healer",
				})
			}
		}

		_, err = datastoreClient.Put(ctx, key, &report)
		if err != nil {
			return fmt.Errorf("datastore write for %s failed: %v", code, err.Error())
		}

		var waitGroup sync.WaitGroup
		var totalErrors uint64
		playerreportTopic := pubsubClient.Topic(playerreportTopicId)
		for _, player := range report.Players {
			result := playerreportTopic.Publish(ctx, &pubsub.Message{
				Attributes: map[string]string{
					"code":      code,
					"player_id": strconv.FormatInt(player.Id, 10),
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
			return fmt.Errorf("%d pubsub writes failed", totalErrors)
		}
	}

	log.Printf("Processed report %v.\n", code)
	return nil
}
