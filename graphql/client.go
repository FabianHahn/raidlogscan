package graphql

import (
	"context"
	"os"

	graphql_lib "github.com/FabianHahn/graphql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	graphqlApiUrl = "https://www.warcraftlogs.com/api/v2/client"
	oauthApiUrl   = "https://www.warcraftlogs.com/oauth"
)

func CreateGraphqlClientOrDie() *graphql_lib.Client {
	config := clientcredentials.Config{
		ClientID:     os.Getenv("WARCRAFTLOGS_CLIENT_ID"),
		ClientSecret: os.Getenv("WARCRAFTLOGS_CLIENT_SECRET"),
		Scopes:       []string{},
		TokenURL:     oauthApiUrl + "/token",
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	oauthClient := config.Client(context.Background())
	return graphql_lib.NewClient(graphqlApiUrl, oauthClient)
}
