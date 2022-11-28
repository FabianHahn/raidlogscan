package graphql

import (
	"context"
	"os"

	graphql_lib "github.com/FabianHahn/graphql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	graphqlApiUrl     = "https://classic.warcraftlogs.com/api/v2/client"
	graphqlUserApiUrl = "https://classic.warcraftlogs.com/api/v2/user"
	oauthApiUrl       = "https://classic.warcraftlogs.com/oauth"
)

func CreateGraphqlClient() *graphql_lib.Client {
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

func CreateGraphqlUserClient(userConfig *oauth2.Config, token *oauth2.Token) *graphql_lib.Client {
	oauthClient := userConfig.Client(context.Background(), token)
	return graphql_lib.NewClient(graphqlUserApiUrl, oauthClient)
}
