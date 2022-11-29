package http

import (
	"context"
	"fmt"
	go_http "net/http"
	"sort"
	"strconv"
	"time"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/datastore"
	"github.com/FabianHahn/raidlogscan/html"
)

func PlayerStats(
	w go_http.ResponseWriter,
	r *go_http.Request,
	datastoreClient *google_datastore.Client,
	accountStatsUrl string,
	claimAccountUrl string,
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
	fmt.Fprintf(w, `<html>
<head>
<title>%v-%v (%v) - WoW Raid Stats</title>
<style type="text/css">
a, a:visited, a:hover, a:active {
	color: inherit;
}

table {
	border-collapse: collapse;
	border: 1px solid black;
}

th {
	border: 1px solid black;
	padding: 3px;
}
td {
	border: 1px solid black;
	padding: 3px;
}

div {
	margin: 10px;
}

.column {
	float: left;
}
</style>
</head>
<body>`, player.Name, player.Server, player.Class)
	fmt.Fprintf(w, "<div>")
	fmt.Fprintf(w, "<h1>%v</h1>\n", player.Name)
	fmt.Fprintf(w, "<b>Class</b>: %v<br>\n", player.Class)
	fmt.Fprintf(w, "<b>Server</b>: %v<br>\n", player.Server)
	if player.Account != "" {
		fmt.Fprintf(w, "<b>Account</b>: <a href=\"%v?account_name=%v\">#%v</a><br>\n", accountStatsUrl, player.Account, player.Account)
	}

	fmt.Fprintf(w, "<form action=\"%v\" method=\"get\">", claimAccountUrl)
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"player_id\" name=\"player_id\" value=\"%v\">", playerId)
	fmt.Fprintf(w, "<label for=\"account_name\"><b>Change account name:</b></label><br>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"account_name\" name=\"account_name\">")
	fmt.Fprintf(w, "&nbsp;<input type=\"submit\" value=\"Change\">")
	fmt.Fprintf(w, "</form>")

	fmt.Fprintf(w, "</div>")

	fmt.Fprintf(w, "<div class=\"column\">")
	fmt.Fprintf(w, "<h2>Coraiders</h2>\n")
	fmt.Fprintf(w, "<table><tr><th>Name</th><th>Count</th></tr>\n")
	for _, entry := range leaderboard {
		if entry.Account == "" {
			if entry.Character.Id == playerId {
				continue
			}

			fmt.Fprintf(w, "<tr><td><a href=\"?player_id=%v\">%v-%v (%v)</a></td><td>%v</td></tr>\n",
				entry.Character.Id,
				entry.Character.Name,
				entry.Character.Server,
				entry.Character.Class,
				entry.Count,
			)
		} else if entry.Account != player.Account {
			fmt.Fprintf(w, "<tr><td><a href=\"%v?account_name=%v\">#%v</a></td><td>%v</td></tr>\n",
				accountStatsUrl,
				entry.Account,
				entry.Account,
				entry.Count,
			)
		}
	}
	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, "</div>")

	fmt.Fprintf(w, "<div class=\"column\">")
	fmt.Fprintf(w, "<h2>Raids</h2>\n")
	fmt.Fprintf(w, "<table><tr><th>Date</th><th>Title</th><th>Zone</th><th>Role</th><th>Spec</th></tr>\n")
	for _, playerReport := range player.Reports {
		if playerReport.Duplicate {
			continue
		}

		fmt.Fprintf(w, "<tr><td>%v</td><td><a href=\"https://classic.warcraftlogs.com/reports/%v\">%v</a></td><td>%v</td><td>%v</td><td>%v</td></tr>\n",
			playerReport.StartTime.Format(time.RFC1123),
			playerReport.Code,
			playerReport.Title,
			playerReport.Zone,
			playerReport.Role,
			playerReport.Spec,
		)
	}
	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, "</div>")
	fmt.Fprintf(w, "</body></html>\n")
}
