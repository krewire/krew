package config

import "github.com/krewire/ui"

// Theme configures the theme switcher and per-mode palette overrides.
type Theme struct {
	// Default is the initial mode: auto, light, or dark.
	Default string `yaml:"default"`
	// Light and Dark hold palette token overrides for each mode.
	Light Palette `yaml:"light"`
	Dark  Palette `yaml:"dark"`
}

// Palette holds color token overrides. Empty tokens fall back to the ui
// defaults.
type Palette struct {
	Base1            string `yaml:"base-1"`
	Base1Content     string `yaml:"base-1-content"`
	Base2            string `yaml:"base-2"`
	Base2Content     string `yaml:"base-2-content"`
	Base3            string `yaml:"base-3"`
	Base3Content     string `yaml:"base-3-content"`
	Primary          string `yaml:"primary"`
	PrimaryContent   string `yaml:"primary-content"`
	Secondary        string `yaml:"secondary"`
	SecondaryContent string `yaml:"secondary-content"`
	Accent           string `yaml:"accent"`
	AccentContent    string `yaml:"accent-content"`
	Ghost            string `yaml:"ghost"`
	GhostContent     string `yaml:"ghost-content"`
	Neutral          string `yaml:"neutral"`
	NeutralContent   string `yaml:"neutral-content"`
	Success          string `yaml:"success"`
	SuccessContent   string `yaml:"success-content"`
	Info             string `yaml:"info"`
	InfoContent      string `yaml:"info-content"`
	Warning          string `yaml:"warning"`
	WarningContent   string `yaml:"warning-content"`
	Error            string `yaml:"error"`
	ErrorContent     string `yaml:"error-content"`
}

// UI merges the palette overrides into the given default palette.
func (p *Palette) UI(defaults ui.Palette) ui.Palette {
	if p == nil {
		return defaults
	}
	if p.Base1 != "" {
		defaults.Base1 = ui.Color(p.Base1)
	}
	if p.Base1Content != "" {
		defaults.Base1Content = ui.Color(p.Base1Content)
	}
	if p.Base2 != "" {
		defaults.Base2 = ui.Color(p.Base2)
	}
	if p.Base2Content != "" {
		defaults.Base2Content = ui.Color(p.Base2Content)
	}
	if p.Base3 != "" {
		defaults.Base3 = ui.Color(p.Base3)
	}
	if p.Base3Content != "" {
		defaults.Base3Content = ui.Color(p.Base3Content)
	}
	if p.Primary != "" {
		defaults.Primary = ui.Color(p.Primary)
	}
	if p.PrimaryContent != "" {
		defaults.PrimaryContent = ui.Color(p.PrimaryContent)
	}
	if p.Secondary != "" {
		defaults.Secondary = ui.Color(p.Secondary)
	}
	if p.SecondaryContent != "" {
		defaults.SecondaryContent = ui.Color(p.SecondaryContent)
	}
	if p.Accent != "" {
		defaults.Accent = ui.Color(p.Accent)
	}
	if p.AccentContent != "" {
		defaults.AccentContent = ui.Color(p.AccentContent)
	}
	if p.Ghost != "" {
		defaults.Ghost = ui.Color(p.Ghost)
	}
	if p.GhostContent != "" {
		defaults.GhostContent = ui.Color(p.GhostContent)
	}
	if p.Neutral != "" {
		defaults.Neutral = ui.Color(p.Neutral)
	}
	if p.NeutralContent != "" {
		defaults.NeutralContent = ui.Color(p.NeutralContent)
	}
	if p.Success != "" {
		defaults.Success = ui.Color(p.Success)
	}
	if p.SuccessContent != "" {
		defaults.SuccessContent = ui.Color(p.SuccessContent)
	}
	if p.Info != "" {
		defaults.Info = ui.Color(p.Info)
	}
	if p.InfoContent != "" {
		defaults.InfoContent = ui.Color(p.InfoContent)
	}
	if p.Warning != "" {
		defaults.Warning = ui.Color(p.Warning)
	}
	if p.WarningContent != "" {
		defaults.WarningContent = ui.Color(p.WarningContent)
	}
	if p.Error != "" {
		defaults.Error = ui.Color(p.Error)
	}
	if p.ErrorContent != "" {
		defaults.ErrorContent = ui.Color(p.ErrorContent)
	}
	return defaults
}
