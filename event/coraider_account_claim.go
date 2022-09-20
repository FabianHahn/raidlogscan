package event

import (
	"context"
	"fmt"
	"log"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/pubsub"
	google_event "github.com/cloudevents/sdk-go/v2/event"
)

func CoraiderAccountClaim(ctx context.Context, e google_event.Event, datastoreClient *google_datastore.Client) error {
	coraiderAccountClaimEvent, err := pubsub.ParseCoraiderAccountClaimEvent(e)
	if err != nil {
		return err
	}

	playerKey := google_datastore.IDKey("player", coraiderAccountClaimEvent.PlayerId, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err.Error())
	}

	var player datastore.Player
	err = tx.Get(playerKey, &player)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(
			"for coraider account claim %v/%v datastore get player %v failed: %v",
			coraiderAccountClaimEvent.ClaimedAccountName,
			coraiderAccountClaimEvent.ClaimedPlayerId,
			coraiderAccountClaimEvent.PlayerId,
			err.Error())
	}

	found := false
	for i := range player.CoraiderAccounts {
		if player.CoraiderAccounts[i].PlayerId == coraiderAccountClaimEvent.ClaimedPlayerId {
			player.CoraiderAccounts[i].Name = coraiderAccountClaimEvent.ClaimedAccountName
			found = true
			break
		}
	}
	if !found {
		player.CoraiderAccounts = append(player.CoraiderAccounts, datastore.PlayerCoraiderAccount{
			PlayerId: coraiderAccountClaimEvent.ClaimedPlayerId,
			Name:     coraiderAccountClaimEvent.ClaimedAccountName,
		})
	}

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		return fmt.Errorf(
			"for coraider account claim %v/%v datastore write player %v failed: %v",
			coraiderAccountClaimEvent.ClaimedAccountName,
			coraiderAccountClaimEvent.ClaimedPlayerId,
			coraiderAccountClaimEvent.PlayerId,
			err.Error())
	}

	_, err = tx.Commit()
	if err != nil {
		return fmt.Errorf(
			"coraider account claim %v/%v player %v datastore transaction failed: %v",
			coraiderAccountClaimEvent.ClaimedAccountName,
			coraiderAccountClaimEvent.ClaimedPlayerId,
			coraiderAccountClaimEvent.PlayerId,
			err.Error())
	}

	log.Printf(
		"Updated coraider account claim %v/%v for player %v.\n",
		coraiderAccountClaimEvent.ClaimedAccountName,
		coraiderAccountClaimEvent.ClaimedPlayerId,
		coraiderAccountClaimEvent.PlayerId)
	return nil
}
