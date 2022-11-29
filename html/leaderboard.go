package html

import "github.com/FabianHahn/raidlogscan/datastore"

type LeaderboardEntry struct {
	Count     int64
	IsAccount bool
	Account   string
	Character datastore.PlayerCoraider
}
