package html

import (
	"fmt"
	"io"
	"time"
)

const guildStatsHtmlTemplate = `{{define "body"}}
<h1>{{.GuildName}}</h1>
<b>Wacraft Logs</b>: <a href="https://classic.warcraftlogs.com/guild/id/{{.GuildId}}">link</a><br>
<b>Raiders</b>: {{len .Leaderboard}}<br>
<b>Raids</b>: {{len .Raids}}<br>
<a href="{{.ScanGuildReportsUrl}}?guild_id={{.GuildId}}">Scan guild raids</a><br>

<div class="column">
  <h2>Raiders</h2>
  <table>
    <tr>
      <th>Name</th>
      <th>Count</th>
    </tr>
{{- range .Leaderboard}}
  {{- if .IsAccount}}
    <tr>
      <td><a href="{{$.AccountStatsUrl}}?account_name={{.Account}}">#{{.Account}}</a></td>
      <td>{{.Count}}</td>
    </tr>
  {{- else}}
    <tr>
      <td><a href="{{$.PlayerStatsUrl}}?player_id={{.Character.Id}}">{{.Character.Name}}-{{.Character.Server}} ({{.Character.Class}})</a></td>
      <td>{{.Count}}</td>
    </tr>
  {{- end}}
{{- end}}
  </table>
</div>

<div class="column">
  <h2>Raids</h2>
  <table>
    <tr>
      <th>Date</th>
      <th>Title</th>
      <th>Zone</th>
      <th>Raiders</th>
    </tr>
{{- range .Raids}}
    <tr>
      <td>{{.StartTime.Format "Mon, 02 Jan 2006 15:04:05 MST"}}</td>
      <td><a href="https://classic.warcraftlogs.com/reports/{{.Code}}">{{.Title}}</a></td>
      <td>{{.Zone}}</td>
      <td>{{.NumPlayers}}</td>
    </tr>
{{- end}}
  </table>
</div>
{{- end}}`

type GuildRaid struct {
	Code       string
	StartTime  time.Time
	Title      string
	Zone       string
	NumPlayers int
}

func (r *Renderer) RenderGuildStats(
	wr io.Writer,
	guildId int32,
	guildName string,
	leaderboard []LeaderboardEntry,
	raids []GuildRaid,
	scanGuildReportsUrl string,
	accountStatsUrl string,
	playerStatsUrl string,
) error {
	return r.templates[guildStatsTemplateName].ExecuteTemplate(wr, baseDefinitionName, struct {
		Title               string
		GuildId             int32
		GuildName           string
		Leaderboard         []LeaderboardEntry
		Raids               []GuildRaid
		ScanGuildReportsUrl string
		AccountStatsUrl     string
		PlayerStatsUrl      string
	}{
		Title:               fmt.Sprintf("%v", guildName),
		GuildId:             guildId,
		GuildName:           guildName,
		Leaderboard:         leaderboard,
		Raids:               raids,
		ScanGuildReportsUrl: scanGuildReportsUrl,
		AccountStatsUrl:     accountStatsUrl,
		PlayerStatsUrl:      playerStatsUrl,
	})
}
