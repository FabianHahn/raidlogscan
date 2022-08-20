package fetchreport

import (
	"context"
	"fmt"
	"log"
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

var datastoreClient *datastore.Client

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
	datastoreClient, err = datastore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	functions.HTTP("FetchReport", fetchReport)
}

// fetchReport is an HTTP Cloud Function.
func fetchReport(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	config := clientcredentials.Config{
		ClientID:     os.Getenv("WARCRAFTLOGS_CLIENT_ID"),
		ClientSecret: os.Getenv("WARCRAFTLOGS_CLIENT_SECRET"),
		Scopes:       []string{},
		TokenURL:     oauthApiUrl + "/token",
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	oauthClient := config.Client(oauth2.NoContext)

	graphqlClient := graphql.NewClient(graphqlApiUrl, oauthClient)
	var query struct {
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
				} `scalar:"true" graphql:"playerDetails(endTime: 999999999999)"`
			} `graphql:"report(code: $code)"`
		}
	}
	variables := map[string]interface{}{
		"code": graphql.String(code),
	}

	err := graphqlClient.Query(ctx, &query, variables)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "GraphQL query failed: %v", err.Error())
		return
	}

	startTime := time.Unix(int64(query.ReportData.Report.StartTime/1000), 0)
	endTime := time.Unix(int64(query.ReportData.Report.EndTime/1000), 0)

	key := datastore.NameKey("report", code, nil)
	_, err = datastoreClient.Put(ctx, key, &query.ReportData.Report)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore write failed: %v", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintf(w, "Title: %v\n", query.ReportData.Report.Title)
	fmt.Fprintf(w, "Start Time: %v\n", startTime.Format(time.RFC1123))
	fmt.Fprintf(w, "End Time: %v\n", endTime.Format(time.RFC1123))
	fmt.Fprintf(w, "Player Details:\n%+v", query.ReportData.Report.PlayerDetails)
}
