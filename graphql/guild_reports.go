package graphql

import (
	"context"
	"fmt"

	graphql_lib "github.com/FabianHahn/graphql"
)

type guildReportsQuery struct {
	ReportData struct {
		Reports struct {
			Data []struct {
				Code graphql_lib.String
			}
			CurrentPage graphql_lib.Int `graphql:"current_page"`
			LastPage    graphql_lib.Int `graphql:"last_page"`
		} `graphql:"reports(guildID: $guildId, page: $page)"`
	}
}

func QueryGuildReports(graphqlClient *graphql_lib.Client, ctx context.Context, guildId int64) ([]string, int, error) {
	var query guildReportsQuery
	page := 1
	reports := []string{}
	for {
		variables := map[string]interface{}{
			"guildId": graphql_lib.Int(guildId),
			"page":    graphql_lib.Int(page),
		}

		err := graphqlClient.Query(ctx, &query, variables)
		if err != nil {
			return reports, 0, fmt.Errorf("GraphQL query failed: %v", err.Error())
		}

		for _, data := range query.ReportData.Reports.Data {
			reports = append(reports, string(data.Code))
		}

		page = int(query.ReportData.Reports.CurrentPage)
		if page < int(query.ReportData.Reports.LastPage) {
			page++
		} else {
			break
		}
	}

	return reports, page, nil
}
