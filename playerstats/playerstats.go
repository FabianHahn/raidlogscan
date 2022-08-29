package playerstats

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

type PlayerReport struct {
	Code      string
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Zone      string
	Spec      string
	Role      string
	Duplicate bool
}

type PlayerCoraider struct {
	Id     int64
	Name   string
	Class  string
	Server string
	Count  int64
}

type PlayerCoraiderAccount struct {
	Name     string
	PlayerId int64
}

type Player struct {
	Name             string
	Class            string
	Server           string
	Account          string
	Reports          []PlayerReport          `datastore:",noindex"`
	Coraiders        []PlayerCoraider        `datastore:",noindex"`
	CoraiderAccounts []PlayerCoraiderAccount `datastore:",noindex"`
	Version          int64
}

type LeaderboardEntry struct {
	Count     int64
	Account   string
	Character PlayerCoraider
}

var datastoreClient *datastore.Client
var accountstatsUrl string
var claimaccountUrl string

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
	datastoreClient, err = datastore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	accountstatsUrl = os.Getenv("RAIDLOGCOUNT_ACCOUNTSTATS_URL")
	claimaccountUrl = os.Getenv("RAIDLOGCOUNT_CLAIMACCOUNT_URL")

	functions.HTTP("PlayerStats", playerStats)
}

// fetchReport is an HTTP Cloud Function.
func playerStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	playerId, err := strconv.ParseInt(r.URL.Query().Get("player_id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Player ID conversion failed: %v", err.Error())
		return
	}

	playerKey := datastore.IDKey("player", playerId, nil)
	var player Player
	err = datastoreClient.Get(ctx, playerKey, &player)
	if err == datastore.ErrNoSuchEntity {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No such player: %v", playerId)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore query failed: %v", err)
		return
	}

	coraiders := map[int64]PlayerCoraider{}
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

	leaderboard := []LeaderboardEntry{}
	for accountName, count := range accountCounts {
		leaderboard = append(leaderboard, LeaderboardEntry{
			Count:   count,
			Account: accountName,
		})
	}

	for _, coraider := range coraiders {
		leaderboard = append(leaderboard, LeaderboardEntry{
			Count: coraider.Count,
			Character: PlayerCoraider{
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
		fmt.Fprintf(w, "<b>Account</b>: <a href=\"%v?account_name=%v\">#%v</a><br>\n", accountstatsUrl, player.Account, player.Account)
	}

	fmt.Fprintf(w, "<form action=\"%v\" method=\"get\">", claimaccountUrl)
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
			fmt.Fprintf(w, "<tr><td><a href=\"?player_id=%v\">%v-%v (%v)</a></td><td>%v</td></tr>\n",
				entry.Character.Id,
				entry.Character.Name,
				entry.Character.Server,
				entry.Character.Class,
				entry.Count,
			)
		} else if entry.Account != player.Account {
			fmt.Fprintf(w, "<tr><td><a href=\"?account_name=%v\">#%v</a></td><td>%v</td></tr>\n",
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
