package datastore

import "time"

type AccountStats struct {
	CreationTime time.Time
	HtmlGzip     []byte `datastore:",noindex"`
}
