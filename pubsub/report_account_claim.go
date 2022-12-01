package pubsub

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"

	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	reportAccountClaimTopicId = "reportaccountclaim"
)

type ReportAccountClaimEvent struct {
	ReportCode         string
	ClaimedPlayerId    int64
	ClaimedAccountName string
}

func ParseReportAccountClaimEvent(e event.Event) (ReportAccountClaimEvent, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return ReportAccountClaimEvent{}, fmt.Errorf("failed to parse event message data: %v", err)
	}

	reportCode := message.Message.Attributes["report_code"]

	claimedPlayerId, err := strconv.ParseInt(message.Message.Attributes["claimed_player_id"], 10, 64)
	if err != nil {
		return ReportAccountClaimEvent{}, fmt.Errorf("claimed Player ID conversion failed: %v", err.Error())
	}

	claimedAccountName := message.Message.Attributes["claimed_account_name"]

	return ReportAccountClaimEvent{
		ReportCode:         reportCode,
		ClaimedPlayerId:    claimedPlayerId,
		ClaimedAccountName: claimedAccountName,
	}, nil
}

func PublishReportAccountClaimEvents(
	pubsubClient *google_pubsub.Client,
	ctx context.Context,
	claimedPlayerId int64,
	claimedAccountName string,
	reportCodes []string,
) error {
	claimedPlayerIdString := strconv.FormatInt(claimedPlayerId, 10)

	var waitGroup sync.WaitGroup
	var totalErrors uint64
	reportTopic := pubsubClient.Topic(reportAccountClaimTopicId)
	for _, reportCode := range reportCodes {
		result := reportTopic.Publish(ctx, &google_pubsub.Message{
			Attributes: map[string]string{
				"report_code":          reportCode,
				"claimed_player_id":    claimedPlayerIdString,
				"claimed_account_name": claimedAccountName,
			},
		})

		waitGroup.Add(1)
		go func(res *google_pubsub.PublishResult) {
			defer waitGroup.Done()
			// The Get method blocks until a server-generated ID or
			// an error is returned for the published message.
			_, err := res.Get(ctx)
			if err != nil {
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

	return nil
}
