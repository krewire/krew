package config

import (
	"strings"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name:   "valid empty config",
			config: Config{},
		},
		{
			name: "valid runtime namespace and env file",
			config: Config{
				Runtime: RuntimeConfig{
					Namespace: "krewire",
					EnvFile:   ".env.local",
					Port:      8080,
				},
			},
		},
		{
			name: "invalid runtime namespace characters",
			config: Config{
				Runtime: RuntimeConfig{
					Namespace: "krewire-app!",
				},
			},
			wantErr: "invalid runtime.namespace",
		},
		{
			name: "invalid runtime env file path traversal",
			config: Config{
				Runtime: RuntimeConfig{
					EnvFile: "../outside.env",
				},
			},
			wantErr: "invalid runtime.env_file",
		},
		{
			name: "invalid port out of bounds",
			config: Config{
				Runtime: RuntimeConfig{
					Port: 70000,
				},
			},
			wantErr: "invalid runtime.port",
		},
		{
			name: "valid multiple services with unique ports",
			config: Config{
				Services: []ServiceConfig{
					{Name: "auth", Port: 8081},
					{Name: "payment", Port: 8082},
				},
			},
		},
		{
			name: "service port conflict (FRK-CONF-032)",
			config: Config{
				Services: []ServiceConfig{
					{Name: "auth", Port: 8081},
					{Name: "payment", Port: 8081},
				},
			},
			wantErr: "port conflict between services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}
