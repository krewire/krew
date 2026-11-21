package commands

import (
	_ "embed"
	"html/template"
)

//go:embed dashboard.html
var dashboardHTML string

// dashboardTemplate is pre-parsed once at initialization to prevent per-request template compilation overhead.
var dashboardTemplate = template.Must(template.New("dashboard").Parse(dashboardHTML))
