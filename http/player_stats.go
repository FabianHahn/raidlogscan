package http

import (
	"context"
	"fmt"
	go_http "net/http"
	"sort"
	"strconv"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/html"
)

func PlayerStats(
	w go_http.ResponseWriter,
	r *go_http.Request,
	htmlRenderer *html.Renderer,
	datastoreClient *google_datastore.Client,
	accountStatsUrl string,
	guildStatsUrl string,
	claimAccountUrl string,
	oauth2LoginUrl string,
) {
	ctx := context.Background()
	playerId, err := strconv.ParseInt(r.URL.Query().Get("player_id"), 10, 64)
	if err != nil {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "Player ID conversion failed: %v", err.Error())
		return
	}

	playerKey := google_datastore.IDKey("player", playerId, nil)
	var player datastore.Player
	err = datastoreClient.Get(ctx, playerKey, &player)
	if err == google_datastore.ErrNoSuchEntity {
		w.WriteHeader(go_http.StatusNotFound)
		fmt.Fprintf(w, "No such player: %v", playerId)
		return
	} else if err != nil {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore query failed: %v", err)
		return
	}

	coraiders := map[int64]datastore.PlayerCoraider{}
	for _, playerCoraider := range player.Coraiders {
		if entry, ok := coraiders[playerCoraider.Id]; ok {
			entry.Count += playerCoraider.Count
			coraiders[playerCoraider.Id] = entry
		} else {
			coraiders[playerCoraider.Id] = playerCoraider
		}
	}

	coaccounts := map[int64]string{}
	for _, playerCoraiderAccount := range player.CoraiderAccounts {
		coaccounts[playerCoraiderAccount.PlayerId] = playerCoraiderAccount.Name
	}

	accountCounts := map[string]int64{}
	for playerId, playerAccountName := range coaccounts {
		if coraider, coraiderExists := coraiders[playerId]; coraiderExists {
			if _, ok := accountCounts[playerAccountName]; !ok {
				accountCounts[playerAccountName] = 0
			}

			accountCounts[playerAccountName] += coraider.Count
			delete(coraiders, playerId)
		}
	}

	leaderboard := []html.LeaderboardEntry{}
	for accountName, count := range accountCounts {
		if accountName == player.Account {
			continue
		}

		leaderboard = append(leaderboard, html.LeaderboardEntry{
			Count:     count,
			IsAccount: true,
			Account:   accountName,
		})
	}

	for _, coraider := range coraiders {
		if coraider.Id == playerId {
			continue
		}

		leaderboard = append(leaderboard, html.LeaderboardEntry{
			Count:     coraider.Count,
			IsAccount: false,
			Character: datastore.PlayerCoraider{
				Id:     coraider.Id,
				Name:   coraider.Name,
				Server: coraider.Server,
				Class:  coraider.Class,
			},
		})
	}
	sort.SliceStable(leaderboard, func(i int, j int) bool {
		return leaderboard[i].Count > leaderboard[j].Count
	})

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	err = htmlRenderer.RenderPlayerStats(
		w,
		playerId,
		player,
		leaderboard,
		accountStatsUrl,
		guildStatsUrl,
		claimAccountUrl,
		oauth2LoginUrl)
	if err != nil {
		fmt.Fprintf(w, "failed to render template: %v", err)
		return
	}
}
