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
	coraiderAccountClaimTopicId = "coraideraccountclaim"
)

type CoraiderAccountClaimEvent struct {
	PlayerId           int64
	ClaimedPlayerId    int64
	ClaimedAccountName string
}

func ParseCoraiderAccountClaimEvent(e event.Event) (CoraiderAccountClaimEvent, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return CoraiderAccountClaimEvent{}, fmt.Errorf("failed to parse event message data: %v", err)
	}

	playerId, err := strconv.ParseInt(message.Message.Attributes["player_id"], 10, 64)
	if err != nil {
		return CoraiderAccountClaimEvent{}, fmt.Errorf("player ID conversion failed: %v", err.Error())
	}
	claimedPlayerId, err := strconv.ParseInt(message.Message.Attributes["claimed_player_id"], 10, 64)
	if err != nil {
		return CoraiderAccountClaimEvent{}, fmt.Errorf("claimed Player ID conversion failed: %v", err.Error())
	}

	claimedAccountName := message.Message.Attributes["claimed_account_name"]

	return CoraiderAccountClaimEvent{
		PlayerId:           playerId,
		ClaimedPlayerId:    claimedPlayerId,
		ClaimedAccountName: claimedAccountName,
	}, nil
}

func PublishCoraiderAccountClaimEvents(
	pubsubClient *google_pubsub.Client,
	ctx context.Context,
	claimedPlayerId int64,
	claimedAccountName string,
	playerIds []int64,
) error {
	claimedPlayerIdString := strconv.FormatInt(claimedPlayerId, 10)

	var waitGroup sync.WaitGroup
	var totalErrors uint64
	reportTopic := pubsubClient.Topic(coraiderAccountClaimTopicId)
	for _, playerId := range playerIds {
		result := reportTopic.Publish(ctx, &google_pubsub.Message{
			Attributes: map[string]string{
				"player_id":            strconv.FormatInt(playerId, 10),
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
