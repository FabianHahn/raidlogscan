package updateplayerreport

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

const (
	graphqlApiUrl = "https://www.warcraftlogs.com/api/v2/client"
	oauthApiUrl   = "https://www.warcraftlogs.com/oauth"
)

type ReportPlayer struct {
	Id     int64
	Name   string
	Class  string
	Server string
	Spec   string
	Role   string
}

type Report struct {
	Title     string
	CreatedAt time.Time
	StartTime time.Time
	EndTime   time.Time
	Players   []ReportPlayer `datastore:",noindex"`
}

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

	functions.HTTP("UpdatePlayerReport", updatePlayerReport)
}

// fetchReport is an HTTP Cloud Function.
func updatePlayerReport(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")
	playerId, err := strconv.ParseInt(r.URL.Query().Get("player_id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Player ID conversion failed: %v", err.Error())
		return
	}

	reportKey := datastore.NameKey("report", code, nil)
	var report Report
	err = datastoreClient.Get(ctx, reportKey, &report)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore report query failed: %v", err.Error())
		return
	}

	var thisReportPlayer *ReportPlayer
	for _, player := range report.Players {
		if player.Id == playerId {
			thisReportPlayer = &player
			break
		}
	}
	if thisReportPlayer == nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Player %v not found in report: %+v", playerId, report)
		return
	}

	playerKey := datastore.IDKey("player", playerId, nil)
	tx, err := datastoreClient.NewTransaction(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to create transaction: %v", err.Error())
		return
	}

	var player Player
	err = tx.Get(playerKey, &player)
	if err == datastore.ErrNoSuchEntity {
		player.Name = thisReportPlayer.Name
		player.Class = thisReportPlayer.Class
		player.Server = thisReportPlayer.Server
	} else if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore get player %v failed: %v", playerKey, err.Error())
		return
	}

	for _, playerReport := range player.Reports {
		if playerReport.Code == code {
			tx.Rollback()
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Report %v already reported for player %v", code, playerKey)
			return
		}
	}

	player.Reports = append(player.Reports, PlayerReport{
		Code:      code,
		Title:     report.Title,
		StartTime: report.StartTime,
		Spec:      thisReportPlayer.Spec,
		Role:      thisReportPlayer.Role,
	})

	coraiders := map[int64]*PlayerCoraider{}
	for id := range player.Coraiders {
		coraider := &player.Coraiders[id]
		coraiders[coraider.Id] = coraider
	}

	currentCoraiders := map[int64]struct{}{}
	for _, reportPlayer := range report.Players {
		if _, alreadyCounted := currentCoraiders[reportPlayer.Id]; alreadyCounted {
			continue
		}

		if coraider, ok := coraiders[reportPlayer.Id]; ok {
			coraider.Count++
		} else {
			coraiders[reportPlayer.Id] = &PlayerCoraider{
				Id:     reportPlayer.Id,
				Name:   reportPlayer.Name,
				Class:  reportPlayer.Class,
				Server: reportPlayer.Server,
				Count:  1,
			}
		}

		currentCoraiders[reportPlayer.Id] = struct{}{}
	}

	player.Coraiders = []PlayerCoraider{}
	for _, coraider := range coraiders {
		player.Coraiders = append(player.Coraiders, *coraider)
	}

	_, err = tx.Put(playerKey, &player)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to update player: %v", err.Error())
		return
	}

	_, err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to commit transaction: %v", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintf(w, "%+v", player)
}
