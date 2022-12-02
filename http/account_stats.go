package http

import (
	"context"
	"fmt"
	"io"
	go_http "net/http"
	"sort"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/cache"
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
	guildStatsUrl string,
	oauth2LoginUrl string,
) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	accountName := r.URL.Query().Get("account_name")
	if accountName == "" {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "No account_name specified")
		return
	}

	accountStatsKey := google_datastore.NameKey("account_stats", accountName, nil)
	var accountStats datastore.AccountStats
	err := datastoreClient.Get(ctx, accountStatsKey, &accountStats)
	if err != nil && err != google_datastore.ErrNoSuchEntity {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore query failed: %v", err)
		return
	} else if err == nil {
		cache.WriteCompressedResponseOrDecompress(w, r, accountStats.HtmlGzip)
		return
	}

	characters := map[int64]datastore.PlayerCoraider{}
	coraiders := map[int64]datastore.PlayerCoraider{}
	guilds := map[int32]html.GuildLeaderboardEntry{}
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

			if playerReport.GuildId != 0 {
				if entry, ok := guilds[playerReport.GuildId]; ok {
					entry.Count++
					guilds[playerReport.GuildId] = entry
				} else {
					guilds[playerReport.GuildId] = html.GuildLeaderboardEntry{
						Count:     1,
						GuildId:   playerReport.GuildId,
						GuildName: playerReport.GuildName,
					}
				}
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

	guildLeaderboard := []html.GuildLeaderboardEntry{}
	for _, guild := range guilds {
		guildLeaderboard = append(guildLeaderboard, guild)
	}
	sort.SliceStable(guildLeaderboard, func(i int, j int) bool {
		return guildLeaderboard[i].Count > guildLeaderboard[j].Count
	})

	cache.CacheAndOutputAccountStats(w, r, datastoreClient, ctx, accountName, func(wr io.Writer) error {
		return htmlRenderer.RenderAccountStats(
			wr,
			accountName,
			numRaids,
			charactersSlice,
			leaderboard,
			guildLeaderboard,
			playerStatsUrl,
			guildStatsUrl,
			oauth2LoginUrl)
	})
}
