package http

import (
	go_http "net/http"

	"github.com/FabianHahn/raidlogscan/oauth2"
	go_oauth2 "golang.org/x/oauth2"
)

const (
	oauthApiUrl = "https://www.warcraftlogs.com/oauth"
	oauthState  = "raidlogscan"
)

func Oauth2Login(
	w go_http.ResponseWriter,
	r *go_http.Request,
	userConfig *go_oauth2.Config,
) {
	url := userConfig.AuthCodeURL(oauth2.Oauth2State)
	go_http.Redirect(w, r, url, go_http.StatusTemporaryRedirect)
}
