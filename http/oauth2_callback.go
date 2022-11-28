package http

import (
	"context"
	"fmt"
	go_http "net/http"

	"github.com/FabianHahn/raidlogscan/graphql"
	go_oauth2 "golang.org/x/oauth2"
)

func Oauth2Callback(
	w go_http.ResponseWriter,
	r *go_http.Request,
	userConfig *go_oauth2.Config,
	scanUserReportsUrl string,
	scanRecentCharacterReportsUrl string,
) {
	ctx := context.Background()

	if r.FormValue("state") != oauthState {
		w.WriteHeader(go_http.StatusForbidden)
		fmt.Fprintf(w, "invalid oauth2 state")
		return
	}

	token, err := userConfig.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		w.WriteHeader(go_http.StatusForbidden)
		fmt.Fprintf(w, "failed oauth2 exchange: %v", err.Error())
		return
	}

	graphqlUserClient := graphql.CreateGraphqlUserClient(userConfig, token)
	userData, err := graphql.QueryUserData(graphqlUserClient, ctx)
	if err != nil {
		w.WriteHeader(go_http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to fetch user data: %v", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, `<html>
<head>
<title>Warcraft Logs Account: %v - WoW Raid Stats</title>
<style type="text/css">
a, a:visited, a:hover, a:active {
	color: inherit;
}

table {
	border-collapse: collapse;
	border: 1px solid black;
}

th {
	border: 1px solid black;
	padding: 3px;
}
td {
	border: 1px solid black;
	padding: 3px;
}

div {
	margin: 10px;
}

.column {
	float: left;
}
</style>
</head>
<body>`, userData.Name)
	fmt.Fprintf(w, "<div>")
	fmt.Fprintf(w, "<h1>Warcraft Logs Account</h1>\n")
	fmt.Fprintf(w, "<b>Account Name</b>: %v<br>\n", userData.Name)
	fmt.Fprintf(w, "<a href=\"%v?user_id=%v\">Scan personal logs</a>\n", scanUserReportsUrl, userData.Id)
	fmt.Fprintf(w, "</div>")

	fmt.Fprintf(w, "<div class=\"column\">")
	fmt.Fprintf(w, "<h2>Characters</h2>\n")
	fmt.Fprintf(w, "<table><tr><th>Name</th><th>Server</th><th>Scan</th></tr>\n")
	for _, character := range userData.Characters {
		fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td><a href=\"%v?character_id=%v\">Scan recent raids</a></td></tr>\n",
			character.Name, character.Server, scanRecentCharacterReportsUrl, character.Id)
	}
	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, "</div>")

	fmt.Fprintf(w, "</body></html>\n")
}
