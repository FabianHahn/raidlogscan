package http

import (
	"context"
	"fmt"
	go_http "net/http"
	"strconv"

	google_datastore "cloud.google.com/go/datastore"
	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/FabianHahn/raidlogscan/pubsub"
)

func ScanGuildReports(
	w go_http.ResponseWriter,
	r *go_http.Request,
	datastoreClient *google_datastore.Client,
	pubsubClient *google_pubsub.Client,
	guildStatsUrl string,
) {
	ctx := context.Background()

	guildId64, err := strconv.ParseInt(r.URL.Query().Get("guild_id"), 10, 32)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "guild ID conversion failed: %v", err.Error())
		return
	}
	guildId := int32(guildId64)

	query := google_datastore.NewQuery("report").FilterField("GuildId", "=", guildId)
	numReports, err := datastoreClient.Count(ctx, query)
	if err != nil {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "Datastore query failed: %v", err)
		return
	}

	if numReports == 0 {
		w.WriteHeader(go_http.StatusForbidden)
		fmt.Fprintf(w, "Can only request guild report scan for a known guild ID with an existing scanned report.")
		return
	}

	err = pubsub.PublishGuildReportsEvent(
		pubsubClient,
		ctx,
		guildId)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "failed to publish guild reports event %v: %v", guildId, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "Successfully requested reports for <a href=\"%v?guild_id=%v\">guild ID %v</a> to be scanned.<br>\n",
		guildStatsUrl,
		guildId,
		guildId,
	)
}
