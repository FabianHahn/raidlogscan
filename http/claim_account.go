package http

import (
	"context"
	"fmt"
	go_http "net/http"
	"strconv"

	google_datastore "cloud.google.com/go/datastore"
	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/pubsub"
)

type PubSubMessage struct {
	Attributes map[string]interface{} `json:"attributes"`
}

type MessagePublishedData struct {
	Message PubSubMessage
}

func ClaimAccount(
	w go_http.ResponseWriter,
	r *go_http.Request,
	datastoreClient *google_datastore.Client,
	pubsubClient *google_pubsub.Client,
	playerStatsUrl string,
	accountStatsUrl string,
) {
	ctx := context.Background()

	accountName := r.URL.Query().Get("account_name")
	playerId, err := strconv.ParseInt(r.URL.Query().Get("player_id"), 10, 64)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "player ID conversion failed: %v", err.Error())
		return
	}

	query := google_datastore.NewQuery("player").FilterField("Name", "=", accountName)
	count, err := datastoreClient.Count(ctx, query)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "player by name %v lookup failed: %v", accountName, err.Error())
		return
	}
	if count == 0 {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "cannot claim account name %v that doesn't correspond to a known character name", accountName)
		return
	}

	playerKey := google_datastore.IDKey("player", playerId, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "failed to create transaction: %v", err.Error())
		return
	}

	var player datastore.Player
	err = tx.Get(playerKey, &player)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "for claim account %v datastore get player %v failed: %v", accountName, playerId, err.Error())
		return
	}

	player.Account = accountName

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "datastore write claim account %v player %v failed: %v", accountName, playerId, err.Error())
		return
	}

	_, err = tx.Commit()
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "datastore write claim account %v player %v failed: %v", accountName, playerId, err.Error())
		return
	}

	coraiderPlayerIds := []int64{}
	for _, coraider := range player.Coraiders {
		coraiderPlayerIds = append(coraiderPlayerIds, coraider.Id)
	}

	err = pubsub.PublishCoraiderAccountClaimEvents(
		pubsubClient,
		ctx,
		playerId,
		accountName,
		coraiderPlayerIds)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "failed to claim account %v player %v: %v", accountName, playerId, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "Successfully assigned character <a href=\"%v?player_id=%v\">%v-%v (%v)</a> to player #<a href=\"%v?account_name=%v\">%v</a>.<br>\n",
		playerStatsUrl,
		playerId,
		player.Name,
		player.Server,
		player.Class,
		accountStatsUrl,
		accountName,
		accountName,
	)
}
