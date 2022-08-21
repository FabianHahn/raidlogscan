package fetchreport

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	graphqlApiUrl = "https://www.warcraftlogs.com/api/v2/client"
	oauthApiUrl   = "https://www.warcraftlogs.com/oauth"
)

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
	Players   []Player `datastore:",noindex"`
}

type ReportQuery struct {
	ReportData struct {
		Report struct {
			Title         graphql.String
			StartTime     graphql.Float
			EndTime       graphql.Float
			PlayerDetails struct {
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
			} `scalar:"true" graphql:"playerDetails(endTime: 999999999999)" datastore:",noindex"`
		} `graphql:"report(code: $code)"`
	}
}

var datastoreClient *datastore.Client
var graphqlClient *graphql.Client

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
	datastoreClient, err = datastore.NewClient(context.Background(), projectID)
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

	oauthClient := config.Client(oauth2.NoContext)
	graphqlClient = graphql.NewClient(graphqlApiUrl, oauthClient)

	functions.HTTP("FetchReport", fetchReport)
}

func convertFloatTime(floatTime float64) time.Time {
	integral, fractional := math.Modf(floatTime / 1000)
	return time.Unix(int64(integral), int64(fractional*1e9))
}

// fetchReport is an HTTP Cloud Function.
func fetchReport(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	key := datastore.NameKey("report", code, nil)
	var report Report
	err := datastoreClient.Get(ctx, key, &report)
	cached := false
	if err == nil {
		cached = true
	} else if err != datastore.ErrNoSuchEntity {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore query failed: %v", err.Error())
		return
	} else {
		var query ReportQuery
		variables := map[string]interface{}{
			"code": graphql.String(code),
		}
		err = graphqlClient.Query(ctx, &query, variables)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "GraphQL query failed: %v", err.Error())
			return
		}

		report.Title = string(query.ReportData.Report.Title)
		report.CreatedAt = time.Now()
		report.StartTime = convertFloatTime(float64(query.ReportData.Report.StartTime))
		report.EndTime = convertFloatTime(float64(query.ReportData.Report.StartTime))
		for _, player := range query.ReportData.Report.PlayerDetails.Data.PlayerDetails.Tanks {
			report.Players = append(report.Players, Player{
				Id:     player.Guid,
				Name:   player.Name,
				Class:  player.Class,
				Server: player.Server,
				Spec:   "",
				Role:   "tank",
			})
		}
		for _, player := range query.ReportData.Report.PlayerDetails.Data.PlayerDetails.Dps {
			report.Players = append(report.Players, Player{
				Id:     player.Guid,
				Name:   player.Name,
				Class:  player.Class,
				Server: player.Server,
				Spec:   player.Spec,
				Role:   "dps",
			})
		}
		for _, player := range query.ReportData.Report.PlayerDetails.Data.PlayerDetails.Healers {
			report.Players = append(report.Players, Player{
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
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Datastore write failed: %v", err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintf(w, "Cached: %v\n", cached)
	fmt.Fprintf(w, "Title: %v\n", report.Title)
	fmt.Fprintf(w, "Created Time: %v\n", report.CreatedAt.Format(time.RFC1123))
	fmt.Fprintf(w, "Start Time: %v\n", report.StartTime.Format(time.RFC1123))
	fmt.Fprintf(w, "End Time: %v\n", report.EndTime.Format(time.RFC1123))
	fmt.Fprintf(w, "Player Details:\n%+v", report.Players)
}
