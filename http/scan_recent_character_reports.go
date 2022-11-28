package http

import (
	"context"
	"fmt"
	go_http "net/http"
	"strconv"

	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/FabianHahn/raidlogscan/pubsub"
)

func ScanRecentCharacterReports(
	w go_http.ResponseWriter,
	r *go_http.Request,
	pubsubClient *google_pubsub.Client,
) {
	ctx := context.Background()

	characterId64, err := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "character ID conversion failed: %v", err.Error())
		return
	}
	characterId := int32(characterId64)

	err = pubsub.PublishRecentCharacterReportsEvent(
		pubsubClient,
		ctx,
		characterId)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "failed to publush user reports event %v: %v", characterId, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "Successfully requested recent reports for character ID %v to be scanned.<br>\n",
		characterId,
	)
}
