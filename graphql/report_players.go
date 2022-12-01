package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	graphql_lib "github.com/FabianHahn/graphql"
)

type ReportPlayers struct {
	Tanks []struct {
		Name   string `json:"name"`
		Guid   int64  `json:"guid"`
		Class  string `json:"type"`
		Server string `json:"server"`
	} `json:"tanks"`
	Dps []struct {
		Name   string `json:"name"`
		Guid   int64  `json:"guid"`
		Class  string `json:"type"`
		Server string `json:"server"`
		Spec   string `json:"icon"`
	} `json:"dps"`
	Healers []struct {
		Name   string `json:"name"`
		Guid   int64  `json:"guid"`
		Class  string `json:"type"`
		Server string `json:"server"`
		Spec   string `json:"icon"`
	} `json:"healers"`
}

type QueryReportResult struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Zone      string
	GuildId   int32
	GuildName string
	Players   ReportPlayers
}

type playerDetailsResponse struct {
	Data struct {
		PlayerDetails ReportPlayers `json:"playerDetails"`
	} `json:"data"`
}

type reportQuery struct {
	ReportData struct {
		Report struct {
			Title     graphql_lib.String
			StartTime graphql_lib.Float
			EndTime   graphql_lib.Float
			Zone      struct {
				Name graphql_lib.String
			}
			Guild struct {
				Id   graphql_lib.Int
				Name graphql_lib.String
			}
			PlayerDetails json.RawMessage `graphql:"playerDetails(endTime: 999999999999)"`
		} `graphql:"report(code: $code)"`
	}
}

func convertFloatTime(floatTime float64) time.Time {
	integral, fractional := math.Modf(floatTime / 1000)
	return time.Unix(int64(integral), int64(fractional*1e9))
}

func QueryReport(graphqlClient *graphql_lib.Client, ctx context.Context, code string) (QueryReportResult, error) {
	result := QueryReportResult{}

	var query reportQuery
	variables := map[string]interface{}{
		"code": graphql_lib.String(code),
	}
	err := graphqlClient.Query(ctx, &query, variables)
	if err != nil {
		return result, fmt.Errorf("GraphQL query for %v failed: %v", code, err.Error())
	}

	result.Title = string(query.ReportData.Report.Title)
	result.StartTime = convertFloatTime(float64(query.ReportData.Report.StartTime))
	result.EndTime = convertFloatTime(float64(query.ReportData.Report.EndTime))
	result.Zone = string(query.ReportData.Report.Zone.Name)
	result.GuildId = int32(query.ReportData.Report.Guild.Id)
	result.GuildName = string(query.ReportData.Report.Guild.Name)

	var playerDetailsResponse playerDetailsResponse
	err = json.Unmarshal(query.ReportData.Report.PlayerDetails, &playerDetailsResponse)
	// The warcraftlogs API is not super stable and sometimes returns empty responses,
	// so failing to parse this is not a disaster, we'll just return an empty response.
	if err == nil {
		result.Players = playerDetailsResponse.Data.PlayerDetails
	}

	return result, nil
}
