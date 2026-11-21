package commands

// Hints maps names developers reach for from other tools to the krew command
// that does that job. `krew <name>` prints the hint instead of a generic
// "unknown command" suggestion.
var Hints = map[string]string{
	"start":     "krew dev  (live dev server with reload) or krew run  (run once)",
	"watch":     "krew dev  (rebuilds and reloads on change)",
	"preview":   "krew serve  (serves the production build; run krew build first)",
	"exec":      "krew run <file.go|task>",
	"create":    "krew new <project>  (new directory) or krew init  (current directory)",
	"install":   "krew add <package> or krew add ui/<component>",
	"i":         "krew add <package>",
	"generate":  "krew make <kind> <name>",
	"g":         "krew make <kind> <name>",
	"component": "krew add ui/<name>  (copy a kui component) or krew make component <Name>  (blank)",
	"lint":      "krew vet  (go vet, krewire.yaml schema, .kiw templates)",
	"check":     "krew vet or krew config",
	"upgrade":   "krew update  (project dependencies) or krew self-update  (the krew CLI)",
	"workspace": "krew dash  (AI-native virtual workspace & sandboxes) or krew ws  (go.work multi-repo)",
	"dashboard": "krew dash  (local dev dashboard, services, logs, sandboxes)",
}
