package http

import (
	"github.com/FabianHahn/raidlogscan/datastore"
)

type LeaderboardEntry struct {
	Count     int64
	Account   string
	Character datastore.PlayerCoraider
}
