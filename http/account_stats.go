package http

import (
	"context"
	"fmt"
	go_http "net/http"
	"sort"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/html"
	"google.golang.org/api/iterator"
)

func AccountStats(
	w go_http.ResponseWriter,
	r *go_http.Request,
	htmlRenderer *html.Renderer,
	datastoreClient *google_datastore.Client,
	playerStatsUrl string,
	oauth2LoginUrl string,
) {
	ctx := context.Background()

	accountName := r.URL.Query().Get("account_name")
	if accountName == "" {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "No account_name specified")
		return
	}

	characters := map[int64]datastore.PlayerCoraider{}
	coraiders := map[int64]datastore.PlayerCoraider{}
	coaccounts := map[int64]string{}
	query := google_datastore.NewQuery("player").FilterField("Account", "=", accountName)
	responseIter := datastoreClient.Run(ctx, query)
	for {
		var player datastore.Player
		key, err := responseIter.Next(&player)
		if err == iterator.Done {
			break
		}
		if err != nil {
			w.WriteHeader(go_http.StatusInternalServerError)
			fmt.Fprintf(w, "Datastore query failed: %v", err)
			return
		}

		character := datastore.PlayerCoraider{
			Id:     key.ID,
			Name:   player.Name,
			Server: player.Server,
			Class:  player.Class,
			Count:  0,
		}

		for _, playerReport := range player.Reports {
			if !playerReport.Duplicate {
				character.Count++
			}
		}

		characters[key.ID] = character

		for _, playerCoraider := range player.Coraiders {
			// Skip our own characters
			if _, ok := characters[playerCoraider.Id]; ok {
				continue
			}

			if entry, ok := coraiders[playerCoraider.Id]; ok {
				entry.Count += playerCoraider.Count
				coraiders[playerCoraider.Id] = entry
			} else {
				coraiders[playerCoraider.Id] = playerCoraider
			}
		}

		for _, playerCoraiderAccount := range player.CoraiderAccounts {
			coaccounts[playerCoraiderAccount.PlayerId] = playerCoraiderAccount.Name
		}
	}

	numRaids := 0
	charactersSlice := []datastore.PlayerCoraider{}
	for _, character := range characters {
		charactersSlice = append(charactersSlice, character)
		numRaids += int(character.Count)
	}
	sort.SliceStable(charactersSlice, func(i int, j int) bool {
		return charactersSlice[i].Count > charactersSlice[j].Count
	})

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
		leaderboard = append(leaderboard, html.LeaderboardEntry{
			Count:     count,
			IsAccount: true,
			Account:   accountName,
		})
	}

	for _, coraider := range coraiders {
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
	err := htmlRenderer.RenderAccountStats(
		w,
		accountName,
		numRaids,
		charactersSlice,
		leaderboard,
		playerStatsUrl,
		oauth2LoginUrl)
	if err != nil {
		fmt.Fprintf(w, "failed to render template: %v", err)
		return
	}
}
