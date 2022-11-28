package graphql

import (
	"context"
	"fmt"

	graphql_lib "github.com/FabianHahn/graphql"
)

type userReportsQuery struct {
	ReportData struct {
		Reports struct {
			Data []struct {
				Code graphql_lib.String
			}
			CurrentPage graphql_lib.Int `graphql:"current_page"`
			LastPage    graphql_lib.Int `graphql:"last_page"`
		} `graphql:"reports(userID: $userId, page: $page)"`
	}
}

func QueryUserReports(graphqlClient *graphql_lib.Client, ctx context.Context, userId int32) ([]string, int, error) {
	var query userReportsQuery
	page := 1
	reports := []string{}
	for {
		variables := map[string]interface{}{
			"userId": graphql_lib.Int(userId),
			"page":   graphql_lib.Int(page),
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
