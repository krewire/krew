package scaffold

import "strings"

func webLayoutsShellTemplate() string {
	return "// Package layouts defines the application page shells (FRK-STR-006).\n" +
		"package layouts\n\n" +
		"import \"github.com/krewire/krewire/pkg/web/ssg\"\n\n" +
		"const shellBody = \"<!DOCTYPE html>\\n\" +\n" +
		"\"<html lang=\\\"en\\\">\\n\" +\n" +
		"\"<head>\\n\" +\n" +
		"\"<meta charset=\\\"utf-8\\\">\\n\" +\n" +
		"\"<meta name=\\\"viewport\\\" content=\\\"width=device-width, initial-scale=1\\\">\\n\" +\n" +
		"\"<meta name=\\\"color-scheme\\\" content=\\\"light dark\\\">\\n\" +\n" +
		"\"<title>{{.Title}}</title>\\n\" +\n" +
		"\"<link rel=\\\"stylesheet\\\" href=\\\"/assets/ui.css\\\">\\n\" +\n" +
		"\"<link rel=\\\"stylesheet\\\" href=\\\"/assets/theme.css\\\">\\n\" +\n" +
		"\"{{themeHead}}\\n\" +\n" +
		"\"</head>\\n\" +\n" +
		"\"<body>\\n\" +\n" +
		"\"<header class=\\\"topbar\\\"><strong>Krewire App</strong><span class=\\\"spacer\\\"></span>{{themeButton}}</header>\\n\" +\n" +
		"\"<main>{{.Content}}</main>\\n\" +\n" +
		"\"<footer class=\\\"footer\\\">Served by a single krewire binary.</footer>\\n\" +\n" +
		"\"</body>\\n\" +\n" +
		"\"</html>\"\n\n" +
		"// Shell is the default application layout: topbar with theme toggle, content,\n" +
		"// footer.\n" +
		"func Shell() ssg.Layout {\n" +
		"\treturn ssg.Layout{\n" +
		"\t\tName:  \"Shell\",\n" +
		"\t\tBody:  shellBody,\n" +
		"\t\tStyle: \".topbar{display:flex;align-items:center;gap:1rem}.spacer{flex:1}\",\n" +
		"\t}\n" +
		"}\n"
}

func webPagesTemplate() string {
	return `// Package pages defines the application's SSR page definitions
// (FRK-STR-006): data + mounts registered at assembly time.
package pages

import (
	"net/http"
	"time"

	"github.com/krewire/krewire/pkg/web"
)

// Register wires all application pages into the web app.
func Register(a *web.App) {
	a.Page(web.PageSpec{
		Path:   "/",
		Title:  "Krewire Monolith",
		Layout: "Shell",
		Root:   "Hero",
		Data: map[string]any{
			"badge":    "v0.1.0",
			"title":    "One model, every mode",
			"subtitle": "SSR, API, and static export from a single runtime.",
			"cta_left": map[string]any{"label": "Read the docs", "href": "/docs"},
		},
		EmitProps: true,
		Scripts:   []string{"/static/app.js"},
	})

	a.Page(web.PageSpec{
		Path:   "/greet",
		Title:  "Greet",
		Layout: "Shell",
		Root:   "Hero",
		DataFunc: func(r *http.Request) (any, error) {
			name := r.URL.Query().Get("name")
			if name == "" {
				name = "world"
			}
			return map[string]any{
				"badge":    time.Now().Format("15:04:05"),
				"title":    "Hello, " + name,
				"subtitle": "This page is rendered per request.",
			}, nil
		},
	})
}
`
}

func webThemeTemplate() string {
	return `// Package theme defines the application palette and toggle styling
// (FRK-STR-006). Palette default: orange neon light/dark, as the ecosystem
// default.
package theme

import "github.com/krewire/krewire/pkg/ui"

// New returns the application theme with the default Krewire palette.
func New() *ui.Theme {
	return &ui.Theme{
		Light: ui.Palette{Primary: "#ff6b00"},
		Dark:  ui.Palette{Primary: "#ff8c42"},
	}
}
`
}

func webComponentsTemplate() string {
	body := strings.ReplaceAll(`// Package components defines the application's SSR components (FRK-STR-006).
// Pages reference these by name via PageSpec.Root.
package components

import "github.com/krewire/krewire/pkg/web/ssg"

// heroBody is the default landing component. It expects a map with the keys
// badge, title, subtitle, and an optional cta_left map of label/href.
const heroBody = @@BT@@<section class="hero">
  {{if .badge}}<span class="hero-badge">{{.badge}}</span>{{end}}
  <h1>{{.title}}</h1>
  {{if .subtitle}}<p class="hero-subtitle">{{.subtitle}}</p>{{end}}
  {{if .cta_left}}<a class="hero-cta" href="{{.cta_left.href}}">{{.cta_left.label}}</a>{{end}}
</section>@@BT@@

// Hero returns the default page component used by the scaffolded routes.
func Hero() ssg.Component {
	return ssg.Component{
		Name:  "Hero",
		Body:  heroBody,
		Style: ".hero{display:grid;gap:.75rem;padding:3rem 0}.hero h1{margin:0;font-size:2.25rem}.hero-subtitle{opacity:.75}.hero-badge{font-size:.75rem;letter-spacing:.08em;text-transform:uppercase;opacity:.6}.hero-cta{margin-top:1rem;padding:.6rem 1.1rem;border-radius:.5rem;background:var(--color-primary,#39D353);color:#000;text-decoration:none;font-weight:600}",
	}
}
`, "@@BT@@", "`")
	return body
}
