package html

import (
	"fmt"
	"io"

	"github.com/FabianHahn/raidlogscan/datastore"
)

const playerStatsHtmlTemplate = `{{define "body"}}
<div>
  <h1>{{.Player.Name}}</h1>
  <b>Class</b>: {{.Player.Class}}<br>
  <b>Server</b>: {{.Player.Server}}<br>
{{- if .HasAccount}}
  <b>Account</b>: <a href="{{.AccountStatsUrl}}?account_name={{.Player.Account}}">#{{.Player.Account}}</a><br>
{{- end}}

  <form action="{{.ClaimAccountUrl}}" method="get">
    <input type="hidden" id="player_id" name="player_id" value="{{.PlayerId}}">
    <label for="account_name"><b>Change account name:</b></label><br>
    <input type="text" id="account_name" name="account_name">&nbsp;
    <input type="submit" value="Change">
  </form>
</div>

<div class="column">
  <h2>Coraiders</h2>
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
      <td><a href="?player_id={{.Character.Id}}">{{.Character.Name}}-{{.Character.Server}} ({{.Character.Class}})</a></td>
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
      <th>Guild</th>
      <th>Zone</th>
      <th>Role</th>
      <th>Spec</th>
    </tr>
{{- range .Player.Reports}}
  {{- if not .Duplicate}}
    <tr>
      <td>{{.StartTime.Format "Mon, 02 Jan 2006 15:04:05 MST"}}</td>
      <td><a href="https://classic.warcraftlogs.com/reports/{{.Code}}">{{.Title}}</a></td>
      <td>
    {{- if ne .GuildId 0}}
        <a href="{{$.GuildStatsUrl}}?guild_id={{.GuildId}}">{{.GuildName}}</a></td>
    {{- end}}
      </td>
      <td>{{.Zone}}</td>
      <td>{{.Role}}</td>
      <td>{{.Spec}}</td>
    </tr>
  {{- end}}
{{- end}}
  </table>
</div>
{{- end}}`

func (r *Renderer) RenderPlayerStats(
	wr io.Writer,
	playerId int64,
	player datastore.Player,
	leaderboard []LeaderboardEntry,
	accountStatsUrl string,
	guildStatsUrl string,
	claimAccountUrl string,
) error {
	return r.templates[playerStatsTemplateName].ExecuteTemplate(wr, baseDefinitionName, struct {
		Title           string
		PlayerId        int64
		Player          datastore.Player
		HasAccount      bool
		Leaderboard     []LeaderboardEntry
		AccountStatsUrl string
		GuildStatsUrl   string
		ClaimAccountUrl string
	}{
		Title:           fmt.Sprintf("%v-%v (%v)", player.Name, player.Server, player.Class),
		PlayerId:        playerId,
		Player:          player,
		HasAccount:      player.Account != "",
		Leaderboard:     leaderboard,
		AccountStatsUrl: accountStatsUrl,
		GuildStatsUrl:   guildStatsUrl,
		ClaimAccountUrl: claimAccountUrl,
	})
}
