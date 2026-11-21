package scaffold

func publicCssTemplate() string {
	return `/* extra example styles */
`
}

func publicJsTemplate() string {
	return `console.log("krewire monolith app loaded");
`
}

func assetsEmbedTemplate() string {
	return `// Package assets embeds the application's static assets served as-is
package assets

import "embed"

//go:embed public
var Files embed.FS
`
}
