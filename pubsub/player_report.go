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
	playerReportTopicId = "playerreport"
)

type PlayerReportEvent struct {
	Code     string
	PlayerId int64
}

func ParsePlayerReportEvent(e event.Event) (PlayerReportEvent, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return PlayerReportEvent{}, fmt.Errorf("failed to parse event message data: %v", err)
	}

	code := message.Message.Attributes["code"]
	playerId, err := strconv.ParseInt(message.Message.Attributes["player_id"], 10, 64)
	if err != nil {
		return PlayerReportEvent{}, fmt.Errorf("player ID conversion failed: %v", err.Error())
	}

	return PlayerReportEvent{
		Code:     code,
		PlayerId: playerId,
	}, nil
}

func PublishPlayerReportEvents(
	pubsubClient *google_pubsub.Client,
	ctx context.Context,
	reportCode string,
	playerIds []int64,
) error {
	var waitGroup sync.WaitGroup
	var totalErrors uint64
	reportTopic := pubsubClient.Topic(playerReportTopicId)
	for _, playerId := range playerIds {
		result := reportTopic.Publish(ctx, &google_pubsub.Message{
			Attributes: map[string]string{
				"code":      reportCode,
				"player_id": strconv.FormatInt(playerId, 10),
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
