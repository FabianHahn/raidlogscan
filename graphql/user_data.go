package graphql

import (
	"context"
	"fmt"

	graphql_lib "github.com/FabianHahn/graphql"
)

type UserDataCharacter struct {
	Id     int32
	Name   string
	Server string
}

type UserDataResult struct {
	Id         int32
	Name       string
	Characters []UserDataCharacter
}

type userDataQuery struct {
	UserData struct {
		CurrentUser struct {
			Name       graphql_lib.String
			Id         graphql_lib.Int
			Characters []struct {
				Id     graphql_lib.Int
				Name   graphql_lib.String
				Server struct {
					Name graphql_lib.String
				}
			}
		}
	}
}

func QueryUserData(graphqlClient *graphql_lib.Client, ctx context.Context) (UserDataResult, error) {
	result := UserDataResult{}

	var query userDataQuery
	variables := map[string]interface{}{}
	err := graphqlClient.Query(ctx, &query, variables)
	if err != nil {
		return result, fmt.Errorf("GraphQL user data query failed: %v", err.Error())
	}

	result.Id = int32(query.UserData.CurrentUser.Id)
	result.Name = string(query.UserData.CurrentUser.Name)
	for _, character := range query.UserData.CurrentUser.Characters {
		result.Characters = append(result.Characters, UserDataCharacter{
			Id:     int32(character.Id),
			Name:   string(character.Name),
			Server: string(character.Server.Name),
		})
	}
	return result, nil
}
