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

	guildId64, err := strconv.ParseInt(r.URL.Query().Get("guild_id"), 10, 64)
	if err != nil {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "Guild ID conversion failed: %v", err.Error())
		return
	}
	guildId := int32(guildId64)

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

		for _, player := range report.Players {
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

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	err = htmlRenderer.RenderGuildStats(
		w,
		guildId,
		guildName,
		leaderboard,
		raids,
		scanGuildReportsUrl,
		accountStatsUrl,
		playerStatsUrl)
	if err != nil {
		fmt.Fprintf(w, "failed to render template: %v", err)
		return
	}
}