package datastore

import "time"

type GuildStats struct {
	CreationTime time.Time
	GuildName    string
	HtmlGzip     []byte `datastore:",noindex"`
}
