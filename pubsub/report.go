package pubsub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	reportTopicId = "report"
)

func ParseReportEvent(e event.Event) (string, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return "", fmt.Errorf("failed to parse event message data: %v", err)
	}
	return message.Message.Attributes["code"], nil
}

func PublishReportEvents(pubsubClient *google_pubsub.Client, ctx context.Context, reports []string) error {
	var waitGroup sync.WaitGroup
	var totalErrors uint64
	reportTopic := pubsubClient.Topic(reportTopicId)
	for _, code := range reports {
		result := reportTopic.Publish(ctx, &google_pubsub.Message{
			Attributes: map[string]string{
				"code": code,
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
