// Package-level file for the static site (VariantStatic):
// krew init --site (or krew new <project> --site) writes a file-based site layout
// with pages/, layouts/, components/, public/, and a clean krewire.yaml.
package scaffold

import (
	"fmt"
)

// equipStatic shapes the bootstrap project into a modern file-based static site. The
// bootstrap's
// placeholder main.go is removed, and standard directories are populated.
func equipStatic(opts EquipOptions) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("equip static: project name is required")
	}
	title := opts.Title
	if title == "" {
		title = opts.Name
	}
	files := []file{
		{krewireYaml, staticKrewireYamlTemplate(opts.Name)},
		{"pages/index.kiw", staticIndexKiwTemplate(title)},
		{"layouts/Base.kiw", staticBaseLayoutTemplate()},
		{"components/Hero.kiw", staticHeroComponentTemplate()},
		{"public/favicon.svg", staticFaviconTemplate()},
		{"README.md", staticReadmeTemplate(opts.Name)},
	}
	report, err := writeVariant(opts.Dir, files)
	if err != nil {
		return nil, err
	}
	if err := removeBootstrapMain(opts.Dir); err != nil {
		return nil, err
	}
	return report, nil
}

func staticKrewireYamlTemplate(name string) string {
	return fmt.Sprintf(`project:
  name: %s
  kind: site
  version: 0.1.0

build:
  output: .krewire/build
  base: /

dev:
  port: 8080
`, name)
}

func staticIndexKiwTemplate(title string) string {
	return fmt.Sprintf(`---
title: %s
layout: Base
---

<div class="hero-container">
  <h1>Welcome to {{.Title}}</h1>
  <p>A fast, modern static site built with Go and Krewire.</p>
  <Hero />
</div>

<style>
  .hero-container {
    max-width: 680px;
    margin: 4rem auto;
    text-align: center;
  }
  .hero-container h1 {
    font-size: 2.75rem;
    color: #39D353;
    margin-bottom: 0.75rem;
    font-weight: 800;
  }
  .hero-container p {
    font-size: 1.15rem;
    color: #8b949e;
    line-height: 1.6;
  }
</style>
`, title)
}

func staticBaseLayoutTemplate() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <link rel="icon" href="/favicon.svg" type="image/svg+xml">
</head>
<body>
  <main>
    {{.Content}}
  </main>
</body>
</html>

<style>
  :root {
    color-scheme: dark;
  }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    margin: 0;
    padding: 2rem;
    background: #0d1117;
    color: #e6edf3;
  }
  a {
    color: #39D353;
    text-decoration: none;
  }
  a:hover {
    text-decoration: underline;
  }
</style>
`
}

func staticHeroComponentTemplate() string {
	return `<div class="card">
  <h3>Getting Started</h3>
  <p>Edit <code>pages/index.kiw</code> to customize this page, or run <code>krew dev</code> to preview changes with instant rebuild.</p>
</div>

<style>
  .card {
    border: 1px solid #30363d;
    border-radius: 8px;
    padding: 1.5rem;
    margin-top: 2rem;
    background: #161b22;
    text-align: left;
  }
  .card h3 {
    margin-top: 0;
    color: #f0f6fc;
  }
  .card p {
    margin-bottom: 0;
    color: #8b949e;
  }
  code {
    background: #21262d;
    padding: 0.2rem 0.4rem;
    border-radius: 4px;
    color: #39D353;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.9em;
  }
</style>
`
}

func staticFaviconTemplate() string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
  <rect width="32" height="32" rx="6" fill="#39D353"/>
  <path d="M16 6L26 24H6L16 6Z" fill="#0d1117"/>
</svg>
`
}

func staticReadmeTemplate(name string) string {
	return fmt.Sprintf(`# %s

A modern static website built with [Krewire](https://krewire.com).

## Development

Start the development server with file watching and auto-rebuild:

`+"```bash"+`
krew dev
`+"```"+`

Visit [http://localhost:8080](http://localhost:8080) to preview your site.

## Building for Production

Compile your site into static assets (`+"`.krewire/build`"+`):

`+"```bash"+`
krew build
`+"```"+`

Preview the production build:

`+"```bash"+`
krew serve
`+"```"+`
`, name)
}
