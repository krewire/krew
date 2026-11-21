package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateHandler(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	pascal := toPascal(name)
	slug := toSlug(name)

	destDir := filepath.Join(root, "internal", "handlers")
	if hasDir(root, "handlers") {
		destDir = filepath.Join(root, "handlers")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + ".go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("handler already exists at %s", rel)
	}

	content := handlerTemplate(pascal, name)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}

func generateMiddleware(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	pascal := toPascal(name)
	slug := toSlug(name)

	destDir := filepath.Join(root, "internal", "middleware")
	if hasDir(root, "middleware") {
		destDir = filepath.Join(root, "middleware")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + ".go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("middleware already exists at %s", rel)
	}

	content := middlewareTemplate(pascal)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}

func generateService(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	pascal := toPascal(name)
	slug := toSlug(name)
	lowerFirst := strings.ToLower(string(pascal[0])) + pascal[1:]

	destDir := filepath.Join(root, "internal", "services")
	if hasDir(root, "services") {
		destDir = filepath.Join(root, "services")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + ".go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("service already exists at %s", rel)
	}

	content := serviceTemplate(pascal, lowerFirst)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}

func generateJob(root, name string) error {
	name = strings.TrimSuffix(name, ".go")
	name = strings.TrimSuffix(name, "Job")
	name = strings.TrimSuffix(name, "job")
	pascal := toPascal(name)
	slug := toSlug(name)

	destDir := filepath.Join(root, "internal", "jobs")
	if hasDir(root, "jobs") {
		destDir = filepath.Join(root, "jobs")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	fileName := strings.ReplaceAll(slug, "-", "_") + "_job.go"
	target := filepath.Join(destDir, fileName)
	if _, err := os.Stat(target); err == nil {
		rel, _ := filepath.Rel(root, target)
		return fmt.Errorf("job already exists at %s", filepath.ToSlash(rel))
	}

	content := jobTemplate(pascal)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, target)
	fmt.Printf("created %s\n", filepath.ToSlash(rel))
	return nil
}
