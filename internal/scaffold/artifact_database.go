package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func generateMigration(root, name string) error {
	_, err := MakeMigration(root, "migrations", name)
	return err
}

// MakeMigration writes a timestamped migration into dir (relative to root)
// and returns its path relative to root. By default, it generates a programmatic
// Go migration (.go) with fluent schema.Builder syntax identical to Laravel.
// If name has a ".sql" extension, it creates a raw SQL migration with "-- +up"
// and "-- +down" sections.
// When another migration already uses the current second, the timestamp moves
// forward so versions stay unique.
func MakeMigration(root, dir, name string) (string, error) {
	isSQL := strings.HasSuffix(name, ".sql")
	name = strings.TrimSuffix(name, ".sql")
	name = strings.TrimSuffix(name, ".go")
	slug := strings.ReplaceAll(toSlug(name), "-", "_")
	if slug == "" {
		return "", fmt.Errorf("migration name %q has no letters or digits", name)
	}
	full := filepath.Join(root, dir)
	if err := os.MkdirAll(full, 0o755); err != nil {
		return "", err
	}

	ext := ".go"
	if isSQL {
		ext = ".sql"
	}

	now := time.Now().UTC()
	for {
		takenGo, err := filepath.Glob(filepath.Join(full, now.Format("20060102150405")+"_*.go"))
		if err != nil {
			return "", err
		}
		takenSQL, err := filepath.Glob(filepath.Join(full, now.Format("20060102150405")+"_*.sql"))
		if err != nil {
			return "", err
		}
		if len(takenGo) == 0 && len(takenSQL) == 0 {
			break
		}
		now = now.Add(time.Second)
	}
	version := now.Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s%s", version, slug, ext)
	target := filepath.Join(full, fileName)

	var content string
	if isSQL {
		content = migrationTemplate(name, now.Format(time.RFC3339))
	} else {
		pkgName := "migrations"
		if base := filepath.Base(dir); base != "" && base != "." {
			cleanBase := strings.ReplaceAll(toSlug(base), "-", "_")
			if cleanBase != "" {
				pkgName = cleanBase
			}
		}

		// Ensure a package entry point exists if dir has no other Go files
		entryFile := filepath.Join(full, pkgName+".go")
		if _, err := os.Stat(entryFile); os.IsNotExist(err) {
			_ = safeWriteArtifact(root, entryFile, []byte(migrationsPackageTemplate(pkgName)))
		}

		content = goMigrationTemplate(pkgName, version, slug)
	}

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return "", err
	}
	rel, _ := filepath.Rel(root, target)
	rel = filepath.ToSlash(rel)
	fmt.Printf("created %s\n", rel)
	return rel, nil
}

func generateModel(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	pascal := toPascal(name)
	slug := toSlug(name)

	destDir := filepath.Join(root, "internal", "models")
	if hasDir(root, "models") {
		destDir = filepath.Join(root, "models")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + ".go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("model already exists at %s", rel)
	}

	content := modelTemplate(pascal)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}

func generateSeeder(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	pascal := toPascal(name)
	if !strings.HasSuffix(pascal, "Seeder") {
		pascal += "Seeder"
	}
	slug := toSlug(pascal)

	destDir := filepath.Join(root, "seeds")
	if hasDir(root, filepath.Join("database", "seeds")) {
		destDir = filepath.Join(root, "database", "seeds")
	} else if hasDir(root, "database") {
		destDir = filepath.Join(root, "database", "seeds")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + ".go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("seeder already exists at %s", rel)
	}

	content := seederTemplate(pascal)
	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}

func generateFactory(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	pascal := toPascal(name)
	if !strings.HasSuffix(pascal, "Factory") {
		pascal += "Factory"
	}
	slug := toSlug(pascal)

	destDir := filepath.Join(root, "database", "factories")
	if !hasDir(root, "database") {
		destDir = filepath.Join(root, "factories")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + ".go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("factory already exists at %s", rel)
	}

	content := factoryTemplate(pascal)
	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}

func generateRepository(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	pascal := toPascal(name)
	if strings.HasSuffix(pascal, "Repository") {
		pascal = strings.TrimSuffix(pascal, "Repository")
	}
	slug := toSlug(pascal) + "_repository"

	destDir := filepath.Join(root, "internal", "repositories")
	if hasDir(root, "repositories") {
		destDir = filepath.Join(root, "repositories")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + ".go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("repository already exists at %s", rel)
	}

	content := repositoryTemplate(pascal)
	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}
