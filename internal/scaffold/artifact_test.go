package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMakeArtifact(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		kind     string
		name     string
		expected string
	}{
		{"component", "Button", filepath.Join("components", "Button.kiw")},
		{"component", "Card.Header", filepath.Join("components", "Card", "Header.kiw")},
		{"component", "Card/Footer", filepath.Join("components", "Card", "Footer.kiw")},
		{"page", "about-us", filepath.Join("pages", "about-us.kiw")},
		{"handler", "User", filepath.Join("internal", "handlers", "user.go")},
		{"model", "Account", filepath.Join("internal", "models", "account.go")},
		{"seeder", "User", filepath.Join("seeds", "user_seeder.go")},
		{"factory", "User", filepath.Join("factories", "user_factory.go")},
		{"repository", "User", filepath.Join("internal", "repositories", "user_repository.go")},
		{"middleware", "Auth", filepath.Join("internal", "middleware", "auth.go")},
		{"service", "Payment", filepath.Join("internal", "services", "payment.go")},
		{"job", "EmailDelivery", filepath.Join("internal", "jobs", "email_delivery_job.go")},
		{"worker", "ReportExport", filepath.Join("internal", "jobs", "report_export_job.go")},
		{"chapter", "getting-started", filepath.Join("content", "01-getting-started.md")},
		{"doc", "api-guide", filepath.Join("content", "02-api-guide.md")},
	}

	for _, tc := range tests {
		t.Run(tc.kind+"_"+tc.name, func(t *testing.T) {
			if err := MakeArtifact(tmpDir, tc.kind, tc.name); err != nil {
				t.Fatalf("MakeArtifact failed for %s %s: %v", tc.kind, tc.name, err)
			}
			target := filepath.Join(tmpDir, tc.expected)
			if _, err := os.Stat(target); os.IsNotExist(err) {
				t.Fatalf("expected file was not created: %s", target)
			}

			// Duplicate test
			if err := MakeArtifact(tmpDir, tc.kind, tc.name); err == nil {
				t.Fatalf("expected error on duplicate creation for %s, got nil", tc.kind)
			}
		})
	}

	// Test migration (Go Laravel-style by default)
	t.Run("migration", func(t *testing.T) {
		if err := MakeArtifact(tmpDir, "migration", "create_users"); err != nil {
			t.Fatalf("MakeArtifact migration failed: %v", err)
		}
		files, err := os.ReadDir(filepath.Join(tmpDir, "migrations"))
		if err != nil || len(files) == 0 {
			t.Fatalf("expected migration file in migrations dir, got %v", files)
		}
		var foundGo bool
		for _, f := range files {
			if strings.HasSuffix(f.Name(), "_create_users.go") {
				foundGo = true
				break
			}
		}
		if !foundGo {
			t.Errorf("migration filename mismatch, expected *_create_users.go in %v", files)
		}
	})

	// Test SQL migration when .sql extension is explicitly provided
	t.Run("migration_sql", func(t *testing.T) {
		if err := MakeArtifact(tmpDir, "migration", "create_orders.sql"); err != nil {
			t.Fatalf("MakeArtifact migration with .sql failed: %v", err)
		}
		files, err := os.ReadDir(filepath.Join(tmpDir, "migrations"))
		if err != nil || len(files) == 0 {
			t.Fatalf("expected migration file in migrations dir, got %v", files)
		}
		var foundSQL bool
		for _, f := range files {
			if strings.HasSuffix(f.Name(), "_create_orders.sql") {
				foundSQL = true
				break
			}
		}
		if !foundSQL {
			t.Errorf("migration filename mismatch, expected *_create_orders.sql in %v", files)
		}
	})

	// Test invalid kind
	t.Run("invalid_kind", func(t *testing.T) {
		if err := MakeArtifact(tmpDir, "unknown_kind", "Test"); err == nil {
			t.Fatal("expected error for unknown kind, got nil")
		}
	})
}

func TestMakeArtifact_PathTraversalSanitization(t *testing.T) {
	tmpDir := t.TempDir()

	// Attempt path traversal when generating an artifact
	err := MakeArtifact(tmpDir, "component", "../../evil")
	if err == nil {
		// If it succeeded, verify it was normalized and contained inside tmpDir/components/
		target := filepath.Join(tmpDir, "components", "Evil.kiw")
		if _, err := os.Stat(target); os.IsNotExist(err) {
			t.Errorf("expected sanitized file inside components/, got: %v", err)
		}
	}

	// Verify nothing escaped outside tmpDir
	parent := filepath.Dir(tmpDir)
	if _, err := os.Stat(filepath.Join(parent, "evil.kiw")); err == nil {
		t.Fatal("CRITICAL: file written outside project root via path traversal!")
	}
}
