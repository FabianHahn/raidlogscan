package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	google_datastore "cloud.google.com/go/datastore"
	"github.com/FabianHahn/raidlogscan/datastore"
)

func InvalidateAccountStatsCache(ctx context.Context, datastoreClient *google_datastore.Client, accountName string) error {
	accountStatsKey := google_datastore.NameKey("account_stats", accountName, nil)
	err := datastoreClient.Delete(ctx, accountStatsKey)
	if err != nil && err != google_datastore.ErrNoSuchEntity {
		return fmt.Errorf("datastore delete failed: %v", err.Error())
	}
	return nil
}

func CacheAndOutputAccountStats(
	w http.ResponseWriter,
	r *http.Request,
	datastoreClient *google_datastore.Client,
	ctx context.Context,
	accountName string,
	render func(wr io.Writer) error,
) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	err := render(writer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to render page: %v", err)
		return
	}

	err = writer.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to gzip compress output: %v", err)
		return
	}

	accountStatsKey := google_datastore.NameKey("account_stats", accountName, nil)
	accountStats := datastore.AccountStats{
		CreationTime: time.Now(),
		HtmlGzip:     buffer.Bytes(),
	}
	_, err = datastoreClient.Put(ctx, accountStatsKey, &accountStats)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to cache account stats: %v", err)
		return
	}

	WriteCompressedResponseOrDecompress(w, r, buffer.Bytes())
}
