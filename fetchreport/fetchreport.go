package fetchreport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	graphqlApiUrl = "https://www.warcraftlogs.com/api/v2/client"
	oauthApiUrl   = "https://www.warcraftlogs.com/oauth"
)

func init() {
	functions.HTTP("FetchReport", fetchReport)
}

// fetchReport is an HTTP Cloud Function.
func fetchReport(w http.ResponseWriter, r *http.Request) {

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
								Name string `json:"name"`
							} `json:"tanks"`
							Dps []struct {
								Name string `json:"name"`
							} `json:"dps"`
							Healers []struct {
								Name string `json:"name"`
							} `json:"healers"`
						} `json:"playerDetails"`
					} `json:"data"`
				} `scalar:"true" graphql:"playerDetails(endTime: 999999999999)"`
			} `graphql:"report(code: $code)"`
		}
	}
	variables := map[string]interface{}{
		"code": graphql.String(r.URL.Query().Get("code")),
	}

	err := graphqlClient.Query(context.Background(), &query, variables)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "GraphQL query failed: %v", err.Error())
		return
	}

	startTime := time.Unix(int64(query.ReportData.Report.StartTime/1000), 0)
	endTime := time.Unix(int64(query.ReportData.Report.EndTime/1000), 0)

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintf(w, "Title: %v\n", query.ReportData.Report.Title)
	fmt.Fprintf(w, "Start Time: %v\n", startTime.Format(time.RFC1123))
	fmt.Fprintf(w, "End Time: %v\n", endTime.Format(time.RFC1123))
	fmt.Fprintf(w, "Player Details:\n%+v", query.ReportData.Report.PlayerDetails)
}
