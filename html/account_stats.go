package html

import (
	"fmt"
	"io"

	"github.com/FabianHahn/raidlogscan/datastore"
)

const accountStatsHtmlTemplate = `{{define "body"}}
<h1>#{{.AccountName}}</h1>
<b>Raids</b>: {{.NumRaids}}<br>
<b>Characters</b>: {{.NumCharacters}}<br>

<div class="column">
  <h2>Coraiders</h2>
  <table>
    <tr>
      <th>Name</th>
      <th>Raids</th>
    </tr>
{{- range .Leaderboard}}
  {{- if .IsAccount}}
    <tr>
      <td><a href="?account_name={{.Account}}">#{{.Account}}</a></td>
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
  <h2>Characters</h2>
  <table>
    <tr>
      <th>Name</th>
      <th>Server</th>
      <th>Class</th>
      <th>Raids</th>
    </tr>
{{- range .Characters}}
    <tr>
      <td><a href="{{$.PlayerStatsUrl}}?player_id={{.Id}}">{{.Name}}</a></td>
      <td>{{.Server}}</td>
      <td>{{.Class}}</td>
      <td>{{.Count}}</td>
    </tr>
{{- end}}
  </table>
  <br>
  Incorrect character assignment? Click<br>
  on the character name and assign them<br>
  a different account name.<br>
  <br>
  Missing character? Find them in a<br>
  raiders list, click on the character<br>
  name and assign them this account<br>
  name ({{.AccountName}}).
</div>

<div class="column">
  <h2>Guilds / Raid Teams</h2>
  <table>
    <tr>
      <th>Name</th>
      <th>Raids</th>
    </tr>
{{- range .GuildLeaderboard}}
    <tr>
      <td><a href="{{$.GuildStatsUrl}}?guild_id={{.GuildId}}">{{.GuildName}}</a></td>
      <td>{{.Count}}</td>
    </tr>
{{- end}}
  </table>
</div>
{{- end}}`

func (r *Renderer) RenderAccountStats(
	wr io.Writer,
	accountName string,
	numRaids int,
	characters []datastore.PlayerCoraider,
	leaderboard []LeaderboardEntry,
	guildLeaderboard []GuildLeaderboardEntry,
	playerStatsUrl string,
	guildStatsUrl string,
	oauth2LoginUrl string,
) error {
	return r.templates[accountStatsTemplateName].ExecuteTemplate(wr, baseDefinitionName, struct {
		Title            string
		AccountName      string
		NumRaids         int
		NumCharacters    int
		Characters       []datastore.PlayerCoraider
		Leaderboard      []LeaderboardEntry
		GuildLeaderboard []GuildLeaderboardEntry
		PlayerStatsUrl   string
		GuildStatsUrl    string
		Oauth2LoginUrl   string
	}{
		Title:            fmt.Sprintf("#%v", accountName),
		AccountName:      accountName,
		NumRaids:         numRaids,
		NumCharacters:    len(characters),
		Characters:       characters,
		Leaderboard:      leaderboard,
		GuildLeaderboard: guildLeaderboard,
		PlayerStatsUrl:   playerStatsUrl,
		GuildStatsUrl:    guildStatsUrl,
		Oauth2LoginUrl:   oauth2LoginUrl,
	})
}
