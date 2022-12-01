package html

import (
	"html/template"
)

const (
	baseDefinitionName       = "base"
	baseTemplateName         = "base.html"
	accountStatsTemplateName = "account_stats.html"
	playerStatsTemplateName  = "player_stats.html"
	guildStatsTemplateName   = "guild_stats.html"
)

type Renderer struct {
	templates map[string]*template.Template
}

func CreateRendererOrDie() *Renderer {
	templates := map[string]*template.Template{}
	templates[accountStatsTemplateName] = template.Must(
		template.Must(
			template.New(accountStatsTemplateName).
				Parse(accountStatsHtmlTemplate)).
			Parse(baseHtmlTemplate))
	templates[playerStatsTemplateName] = template.Must(
		template.Must(
			template.New(playerStatsTemplateName).
				Parse(playerStatsHtmlTemplate)).
			Parse(baseHtmlTemplate))
	templates[guildStatsTemplateName] = template.Must(
		template.Must(
			template.New(guildStatsTemplateName).
				Parse(guildStatsHtmlTemplate)).
			Parse(baseHtmlTemplate))
	return &Renderer{
		templates: templates,
	}
}
