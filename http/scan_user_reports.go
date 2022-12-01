package http

import (
	"context"
	"fmt"
	go_http "net/http"
	"strconv"

	google_pubsub "cloud.google.com/go/pubsub"
	"github.com/FabianHahn/raidlogscan/pubsub"
)

func ScanUserReports(
	w go_http.ResponseWriter,
	r *go_http.Request,
	pubsubClient *google_pubsub.Client,
) {
	ctx := context.Background()

	userId64, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 32)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "user ID conversion failed: %v", err.Error())
		return
	}
	userId := int32(userId64)

	err = pubsub.PublishUserReportsEvent(
		pubsubClient,
		ctx,
		userId)
	if err != nil {
		w.WriteHeader(go_http.StatusBadRequest)
		fmt.Fprintf(w, "failed to publish user reports event %v: %v", userId, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "Successfully requested reports for user ID %v to be scanned.<br>\n",
		userId,
	)
}
