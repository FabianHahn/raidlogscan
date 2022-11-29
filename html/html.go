package html

import (
	"html/template"
)

const (
	baseDefinitionName       = "base"
	baseTemplateName         = "base.html"
	accountStatsTemplateName = "account_stats.html"
)

type Renderer struct {
	templates map[string]*template.Template
}

func CreateRendererOrDie() *Renderer {
	templates := map[string]*template.Template{}
	templates[accountStatsTemplateName] = template.Must(
		template.Must(
			template.New(accountStatsTemplateName).Parse(accountStatsHtmlTemplate)).Parse(baseHtmlTemplate))
	return &Renderer{
		templates: templates,
	}
}
