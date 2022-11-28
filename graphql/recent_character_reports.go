package graphql

import (
	"context"
	"fmt"

	graphql_lib "github.com/FabianHahn/graphql"
)

type recentCharacterReportsQuery struct {
	CharacterData struct {
		Character struct {
			RecentReports struct {
				Data []struct {
					Code graphql_lib.String
				}
				CurrentPage graphql_lib.Int `graphql:"current_page"`
				LastPage    graphql_lib.Int `graphql:"last_page"`
			} `graphql:"recentReports(page: $page)"`
		} `graphql:"character(id: $characterId)"`
	}
}

func QueryRecentCharacterReports(graphqlClient *graphql_lib.Client, ctx context.Context, characterId int32) ([]string, int, error) {
	var query recentCharacterReportsQuery
	page := 1
	reports := []string{}
	for {
		variables := map[string]interface{}{
			"characterId": graphql_lib.Int(characterId),
			"page":        graphql_lib.Int(page),
		}

		err := graphqlClient.Query(ctx, &query, variables)
		if err != nil {
			return reports, 0, fmt.Errorf("GraphQL query failed: %v", err.Error())
		}

		for _, data := range query.CharacterData.Character.RecentReports.Data {
			reports = append(reports, string(data.Code))
		}

		page = int(query.CharacterData.Character.RecentReports.CurrentPage)
		if page < int(query.CharacterData.Character.RecentReports.LastPage) {
			page++
		} else {
			break
		}
	}

	return reports, page, nil
}
