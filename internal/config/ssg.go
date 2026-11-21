package config

import (
	"github.com/krewire/krewire/pkg/ui"
	"github.com/krewire/krewire/pkg/web/ssg"
)

// SSGConfig holds SSG-specific fields that live under the `ssg:` key in
// krewire.yaml. It mirrors the fields of ssg.Config that are not shared with
// the top-level Config.
type SSGConfig struct {
	// Description is injected into every page's template data.
	Description string `yaml:"description"`
	// Layouts are the site's named layouts.
	Layouts []SSGLayoutConfig `yaml:"layouts"`
	// Components are the site's named components.
	Components []SSGComponentConfig `yaml:"components"`
	// Pages are the site's output pages.
	Pages []SSGPageConfig `yaml:"pages"`
	// Assets maps output paths to file contents.
	Assets map[string]string `yaml:"assets"`
	// AutoAssets carries the ssg.auto_assets block into the ssg.Config.
	AutoAssets *ssg.AutoAssetConfig `yaml:"auto_assets"`
}

// SSGLayoutConfig is a layout entry under ssg.layouts.
type SSGLayoutConfig struct {
	Name  string `yaml:"name"`
	Body  string `yaml:"body"`
	Style string `yaml:"style"`
	// UI, when set, builds the layout from a reusable pkg/ui shell
	// instead of Body/Style.
	UI *ui.LayoutConfig `yaml:"ui"`
}

// SSGComponentConfig is a component entry under ssg.components.
type SSGComponentConfig struct {
	Name  string `yaml:"name"`
	Body  string `yaml:"body"`
	Style string `yaml:"style"`
}

// SSGPageConfig is a page entry under ssg.pages.
type SSGPageConfig struct {
	Path   string         `yaml:"path"`
	Title  string         `yaml:"title"`
	Layout string         `yaml:"layout"`
	Root   string         `yaml:"root"`
	Data   map[string]any `yaml:"data"`
}

// ToSSGConfig converts this Config into an ssg.Config, merging top-level
// fields (Title, Base, Output, Theme) with the SSG-specific fields under ssg:.
// Callers must check IsSSG() before calling this method.
func (c *Config) ToSSGConfig() *ssg.Config {
	s := c.SSG
	autoAssets := s.AutoAssets
	if autoAssets == nil {
		autoAssets = c.AutoAssets
	}
	config := &ssg.Config{
		Title:       c.Title,
		Description: s.Description,
		Base:        c.Base,
		Version:     c.Version,
		Output:      c.Output,
		Assets:      s.Assets,
		AutoAssets:  autoAssets,
	}
	if c.Theme != nil {
		config.Theme = &ssg.ThemeConfig{
			Default: c.Theme.Default,
		}
	}
	for _, l := range s.Layouts {
		config.Layouts = append(config.Layouts, ssg.LayoutConfig{Name: l.Name, Body: l.Body, Style: l.Style, UI: l.UI})
	}
	for _, comp := range s.Components {
		config.Components = append(config.Components, ssg.ComponentConfig{Name: comp.Name, Body: comp.Body, Style: comp.Style})
	}
	for _, p := range s.Pages {
		config.Pages = append(config.Pages, ssg.PageConfig{
			Path:   p.Path,
			Title:  p.Title,
			Layout: p.Layout,
			Root:   p.Root,
			Data:   p.Data,
		})
	}
	return config
}
