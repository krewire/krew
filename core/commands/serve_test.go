package commands

import (
	"flag"
	"testing"

	"github.com/krewire/krew/core/config"
)

func TestRegisterServeFlags(t *testing.T) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	RegisterServe(fs)

	expectedFlags := []string{"input", "addr", "port", "output", "o", "env", "debug"}
	for _, name := range expectedFlags {
		if fs.Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered in RegisterServe", name)
		}
	}
}

func TestResolveAddrWithPort(t *testing.T) {
	tests := []struct {
		name     string
		flagAddr string
		flagPort string
		cfgPort  int
		want     string
	}{
		{
			name: "default",
			want: ":8080",
		},
		{
			name:     "flag port number",
			flagPort: "3000",
			want:     ":3000",
		},
		{
			name:     "flag port colon",
			flagPort: ":4000",
			want:     ":4000",
		},
		{
			name:     "flag addr full",
			flagAddr: "127.0.0.1:9090",
			want:     "127.0.0.1:9090",
		},
		{
			name:     "flag addr port only",
			flagAddr: "8000",
			want:     ":8000",
		},
		{
			name:     "flag addr overrides flag port",
			flagAddr: ":7070",
			flagPort: "8000",
			want:     ":7070",
		},
		{
			name:    "config declared port fallback",
			cfgPort: 5000,
			want:    ":5000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *config.Config
			if tc.cfgPort > 0 {
				cfg = &config.Config{}
				cfg.Runtime.Port = tc.cfgPort
			}
			got := resolveAddrWithPort(tc.flagAddr, tc.flagPort, cfg)
			if got != tc.want {
				t.Errorf("resolveAddrWithPort(%q, %q) = %q; want %q", tc.flagAddr, tc.flagPort, got, tc.want)
			}
		})
	}
}
