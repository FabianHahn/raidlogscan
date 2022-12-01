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
	guildReportsTopicId = "guildreports"
)

func ParseGuildReportsEvent(e event.Event) (int64, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return 0, fmt.Errorf("failed to parse event message data: %v", err)
	}
	return strconv.ParseInt(message.Message.Attributes["guild_id"], 10, 64)
}

func PublishGuildReportsEvent(
	pubsubClient *google_pubsub.Client,
	ctx context.Context,
	guildId int32,
) error {
	var waitGroup sync.WaitGroup
	var totalErrors uint64
	reportTopic := pubsubClient.Topic(guildReportsTopicId)
	result := reportTopic.Publish(ctx, &google_pubsub.Message{
		Attributes: map[string]string{
			"guild_id": strconv.FormatInt(int64(guildId), 10),
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
