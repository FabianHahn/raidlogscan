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
	recentCharacterReportsTopicId = "recentcharacterreports"
)

func ParseRecentCharacterReportsEvent(e event.Event) (int32, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return 0, fmt.Errorf("failed to parse event message data: %v", err)
	}
	userId64, err := strconv.ParseInt(message.Message.Attributes["character_id"], 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(userId64), nil
}

func PublishRecentCharacterReportsEvent(
	pubsubClient *google_pubsub.Client,
	ctx context.Context,
	userId int32,
) error {
	var waitGroup sync.WaitGroup
	var totalErrors uint64
	reportTopic := pubsubClient.Topic(recentCharacterReportsTopicId)
	result := reportTopic.Publish(ctx, &google_pubsub.Message{
		Attributes: map[string]string{
			"character_id": strconv.FormatInt(int64(userId), 10),
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
	waitGroup.Wait()

	if totalErrors > 0 {
		return fmt.Errorf("%d pubsub writes failed", totalErrors)
	}

	return nil
}
