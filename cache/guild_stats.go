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

func InvalidateGuildStatsCache(ctx context.Context, datastoreClient *google_datastore.Client, guildId int32) error {
	guildId64 := int64(guildId)
	guildStatsKey := google_datastore.IDKey("guild_stats", guildId64, nil)
	err := datastoreClient.Delete(ctx, guildStatsKey)
	if err != nil && err != google_datastore.ErrNoSuchEntity {
		return fmt.Errorf("datastore delete failed: %v", err.Error())
	}
	return nil
}

func CacheAndOutputGuildStats(
	w http.ResponseWriter,
	r *http.Request,
	datastoreClient *google_datastore.Client,
	ctx context.Context,
	guildId int32,
	guildName string,
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

	guildId64 := int64(guildId)
	guildStatsKey := google_datastore.IDKey("guild_stats", guildId64, nil)
	guildStats := datastore.GuildStats{
		CreationTime: time.Now(),
		GuildName:    guildName,
		HtmlGzip:     buffer.Bytes(),
	}
	_, err = datastoreClient.Put(ctx, guildStatsKey, &guildStats)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to cache guild stats: %v", err)
		return
	}

	WriteCompressedResponseOrDecompress(w, r, buffer.Bytes())
}
