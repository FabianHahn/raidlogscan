package accountstats

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
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
var playerstatsUrl string

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
	datastoreClient, err = datastore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

	playerstatsUrl = os.Getenv("RAIDLOGCOUNT_PLAYERSTATS_URL")

	functions.HTTP("AccountStats", accountStats)
}

// fetchReport is an HTTP Cloud Function.
func accountStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	accountName := r.URL.Query().Get("account_name")
	if accountName == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No account_name specified")
		return
	}

	characters := map[int64]PlayerCoraider{}
	coraiders := map[int64]PlayerCoraider{}
	coaccounts := map[int64]string{}
	query := datastore.NewQuery("player").FilterField("Account", "=", accountName)
	responseIter := datastoreClient.Run(ctx, query)
	for {
		var player Player
		key, err := responseIter.Next(&player)
		if err == iterator.Done {
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Datastore query failed: %v", err)
			return
		}

		character := PlayerCoraider{
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
	charactersSlice := []PlayerCoraider{}
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
<body>`)
	fmt.Fprintf(w, "<div>")
	fmt.Fprintf(w, "<h1>#%v</h1>\n", accountName)
	fmt.Fprintf(w, "<b>Raids</b>: %v<br>\n", numRaids)
	fmt.Fprintf(w, "<b>Characters</b>: %v<br>\n", len(charactersSlice))

	fmt.Fprintf(w, "<div class=\"column\">")
	fmt.Fprintf(w, "<h2>Coraiders</h2>\n")
	fmt.Fprintf(w, "<table><tr><th>Name</th><th>Count</th></tr>\n")
	for _, entry := range leaderboard {
		if entry.Account == "" {
			fmt.Fprintf(w, "<tr><td><a href=\"%v?player_id=%v\">%v-%v (%v)</a></td><td>%v</td></tr>\n",
				playerstatsUrl,
				entry.Character.Id,
				entry.Character.Name,
				entry.Character.Server,
				entry.Character.Class,
				entry.Count,
			)
		} else {
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
	fmt.Fprintf(w, "<h2>Characters</h2>\n")
	fmt.Fprintf(w, "<table><tr><th>Name</th><th>Server</th><th>Class</th><th>Count</th></tr>\n")
	for _, character := range charactersSlice {
		fmt.Fprintf(w, "<tr><td><a href=\"%v?player_id=%v\">%v</a></td><td>%v</td><td>%v</td><td>%v</td></tr>\n",
			playerstatsUrl,
			character.Id,
			character.Name,
			character.Server,
			character.Class,
			character.Count,
		)
	}
	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, "</div>\n")

	fmt.Fprintf(w, "</body></html>\n")
}
