package oauth2

import (
	"os"

	"golang.org/x/oauth2"
)

const (
	Oauth2State  = "raidlogscan"
	oauth2ApiUrl = "https://classic.warcraftlogs.com/oauth"
)

func CreateOauth2UserConfig(redirectUrl string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("WARCRAFTLOGS_CLIENT_ID"),
		ClientSecret: os.Getenv("WARCRAFTLOGS_CLIENT_SECRET"),
		RedirectURL:  redirectUrl,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:   oauth2ApiUrl + "/authorize",
			TokenURL:  oauth2ApiUrl + "/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
}
