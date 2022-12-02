package http

import (
	"context"
	"fmt"
	"io"
	go_http "net/http"
	"sort"
	"strconv"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/cache"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/html"
	"google.golang.org/api/iterator"
)

func GuildStats(
	w go_http.ResponseWriter,
	r *go_http.Request,
	htmlRenderer *html.Renderer,
	datastoreClient *google_datastore.Client,
	scanGuildReportsUrl string,
	accountStatsUrl string,
	playerStatsUrl string,
) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	guildId64, err := strconv.ParseInt(r.URL.Query().Get("guild_id"), 10, 64)
	if err != nil {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "Guild ID conversion failed: %v", err.Error())
		return
	}
	guildId := int32(guildId64)

	guildStatsKey := google_datastore.IDKey("guild_stats", guildId64, nil)
	var guildStats datastore.GuildStats
	err = datastoreClient.Get(ctx, guildStatsKey, &guildStats)
	if err != nil && err != google_datastore.ErrNoSuchEntity {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore query failed: %v", err)
		return
	} else if err == nil {
		cache.WriteCompressedResponseOrDecompress(w, r, guildStats.HtmlGzip)
		return
	}

	guildName := ""
	raids := []html.GuildRaid{}
	playerAccounts := map[int64]string{}
	accountCounts := map[string]int64{}
	raiders := map[int64]datastore.PlayerCoraider{}
	query := google_datastore.NewQuery("report").FilterField("GuildId", "=", guildId).Order("-StartTime")
	responseIter := datastoreClient.Run(ctx, query)
	for {
		var report datastore.Report
		key, err := responseIter.Next(&report)
		if err == iterator.Done {
			break
		}
		if err != nil {
			w.WriteHeader(go_http.StatusInternalServerError)
			fmt.Fprintf(w, "Datastore query failed: %v", err)
			return
		}

		guildName = report.GuildName
		raids = append(raids, html.GuildRaid{
			Code:       key.Name,
			StartTime:  report.StartTime,
			Title:      report.Title,
			Zone:       report.Zone,
			NumPlayers: len(report.Players),
		})

		for _, playerAccount := range report.PlayerAccounts {
			playerAccounts[playerAccount.PlayerId] = playerAccount.Name
		}

		reportPlayers := map[int64]struct{}{}
		for _, player := range report.Players {
			// Don't count duplicate players in a report multiple times
			if _, ok := reportPlayers[player.Id]; ok {
				continue
			}
			reportPlayers[player.Id] = struct{}{}

			if accountName, ok := playerAccounts[player.Id]; ok {
				accountCounts[accountName]++
				continue
			}

			if entry, ok := raiders[player.Id]; ok {
				entry.Count += 1
				raiders[player.Id] = entry
			} else {
				raiders[player.Id] = datastore.PlayerCoraider{
					Id:     player.Id,
					Name:   player.Name,
					Server: player.Server,
					Class:  player.Class,
					Count:  1,
				}
			}
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

	for _, raider := range raiders {
		leaderboard = append(leaderboard, html.LeaderboardEntry{
			Count:     raider.Count,
			IsAccount: false,
			Character: datastore.PlayerCoraider{
				Id:     raider.Id,
				Name:   raider.Name,
				Server: raider.Server,
				Class:  raider.Class,
			},
		})
	}
	sort.SliceStable(leaderboard, func(i int, j int) bool {
		return leaderboard[i].Count > leaderboard[j].Count
	})

	cache.CacheAndOutputGuildStats(w, r, datastoreClient, ctx, guildId, guildName, func(wr io.Writer) error {
		return htmlRenderer.RenderGuildStats(
			wr,
			guildId,
			guildName,
			leaderboard,
			raids,
			scanGuildReportsUrl,
			accountStatsUrl,
			playerStatsUrl)
	})
}
