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
	Spec      string
	Role      string
}

type PlayerCoraider struct {
	Id     int64
	Name   string
	Class  string
	Server string
	Count  int64
}

type Player struct {
	Name      string
	Class     string
	Server    string
	Reports   []PlayerReport   `datastore:",noindex"`
	Coraiders []PlayerCoraider `datastore:",noindex"`
}

var datastoreClient *datastore.Client

func init() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
	var err error
	datastoreClient, err = datastore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatal(err)
	}

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

	sort.SliceStable(player.Reports, func(i int, j int) bool {
		return player.Reports[i].StartTime.After(player.Reports[j].StartTime)
	})
	sort.SliceStable(player.Coraiders, func(i int, j int) bool {
		return player.Coraiders[i].Count > player.Coraiders[j].Count
	})

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, `<html>
<head>
<style type="text/css">
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
	fmt.Fprintf(w, "<h1>%v</h1>\n", player.Name)
	fmt.Fprintf(w, "<b>Class</b>: %v<br>\n", player.Class)
	fmt.Fprintf(w, "<b>Server</b>: %v<br>\n", player.Server)
	fmt.Fprintf(w, "</div>")
	fmt.Fprintf(w, "<div class=\"column\">")
	fmt.Fprintf(w, "<h2>Coraiders</h2>\n")
	fmt.Fprintf(w, "<table><tr><th>Name</th><th>Class</th><th>Count</th></tr>\n")
	for _, playerCoraider := range player.Coraiders {
		fmt.Fprintf(w, "<tr><td><a href=\"?player_id=%v\">%v</a></td><td>%v</td><td>%v</td></tr>\n",
			playerCoraider.Id,
			playerCoraider.Name,
			playerCoraider.Class,
			playerCoraider.Count,
		)
	}
	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, "</div>")
	fmt.Fprintf(w, "<div class=\"column\">")
	fmt.Fprintf(w, "<h2>Raids</h2>\n")
	fmt.Fprintf(w, "<table><tr><th>Date</th><th>Title</th><th>Role</th><th>Spec</th></tr>\n")
	for _, playerReport := range player.Reports {
		fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>\n",
			playerReport.StartTime.Format(time.RFC1123),
			playerReport.Title,
			playerReport.Role,
			playerReport.Spec,
		)
	}
	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, "</div>")
	fmt.Fprintf(w, "</body></html>\n")
}
