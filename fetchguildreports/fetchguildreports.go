package fetchguildreports

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	graphqlApiUrl = "https://www.warcraftlogs.com/api/v2/client"
	oauthApiUrl   = "https://www.warcraftlogs.com/oauth"
	reportTopicId = "report"
)

type MessagePublishedData struct {
	Message pubsub.Message
}

type GuildReportsQuery struct {
	ReportData struct {
		Reports struct {
			Data []struct {
				Code graphql.String
			}
			CurrentPage graphql.Int `graphql:"current_page"`
			LastPage    graphql.Int `graphql:"last_page"`
		} `graphql:"reports(guildID: $guildId, page: $page)"`
	}
}

var pubsubClient *pubsub.Client
var graphqlClient *graphql.Client

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
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

	functions.CloudEvent("FetchGuildReports", fetchGuildReports)
}

// fetchReport is an HTTP Cloud Function.
func fetchGuildReports(ctx context.Context, e event.Event) error {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return fmt.Errorf("failed to parse event message data: %v", err)
	}
	guildId, err := strconv.ParseInt(message.Message.Attributes["guild_id"], 10, 64)
	if err != nil {
		return fmt.Errorf("guild ID conversion failed: %v", err.Error())
	}

	var query GuildReportsQuery
	page := 1
	reports := []string{}
	for {
		variables := map[string]interface{}{
			"guildId": graphql.Int(guildId),
			"page":    graphql.Int(page),
		}

		err = graphqlClient.Query(ctx, &query, variables)
		if err != nil {
			return fmt.Errorf("GraphQL query failed: %v", err.Error())
		}

		for _, data := range query.ReportData.Reports.Data {
			reports = append(reports, string(data.Code))
		}

		page = int(query.ReportData.Reports.CurrentPage)
		if page < int(query.ReportData.Reports.LastPage) {
			page++
		} else {
			break
		}
	}

	var waitGroup sync.WaitGroup
	var totalErrors uint64
	reportTopic := pubsubClient.Topic(reportTopicId)
	for _, code := range reports {
		result := reportTopic.Publish(ctx, &pubsub.Message{
			Attributes: map[string]string{
				"code": code,
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
				log.Printf("Failed to publish: %v", err)
				atomic.AddUint64(&totalErrors, 1)
				return
			}
		}(result)
	}
	waitGroup.Wait()

	if totalErrors > 0 {
		return fmt.Errorf("%d pubsub writes failed", totalErrors)
	}

	log.Printf("Fetched %v reports in %v pages.\n", len(reports), page)
	return nil
}
