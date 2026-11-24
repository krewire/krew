package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func generateChapter(root, name string) error {
	name = strings.TrimSuffix(name, ".md")
	title := toTitle(name)
	slug := toSlug(name)

	// Determine destination directory
	destDir := "content"
	if hasDir(root, filepath.Join("content", "docs")) {
		destDir = filepath.Join("content", "docs")
	} else if hasDir(root, filepath.Join("content", "en", "docs")) {
		destDir = filepath.Join("content", "en", "docs")
	}

	fullDir := filepath.Join(root, destDir)
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		return err
	}

	// Check if any existing file already has the same slug or name
	entries, _ := os.ReadDir(fullDir)
	mdCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			mdCount++
			base := strings.TrimSuffix(e.Name(), ".md")
			parts := strings.SplitN(base, "-", 2)
			if len(parts) == 2 && isNumeric(parts[0]) {
				base = parts[1]
			}
			if toSlug(base) == slug {
				return fmt.Errorf("chapter %q already exists at %s", name, filepath.Join(destDir, e.Name()))
			}
		}
	}
	nextOrder := mdCount + 1

	// File name formatting: if user didn't include numbers, prefix with 01-, 02-, etc.
	fileName := slug + ".md"
	if len(slug) > 0 && !unicode.IsDigit(rune(slug[0])) {
		fileName = fmt.Sprintf("%02d-%s.md", nextOrder, slug)
	}

	target := filepath.Join(fullDir, fileName)
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("chapter %q already exists at %s", name, filepath.Join(destDir, fileName))
	}

	content := chapterTemplate(title, nextOrder)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	fmt.Printf("created %s\n", filepath.ToSlash(filepath.Join(destDir, fileName)))
	return nil
}

func generateComponent(root, name string) error {
	name = strings.TrimSuffix(name, ".kiw")
	name = strings.ReplaceAll(name, "\\", "/")

	// Support compound and nested components e.g. "Card.Header" or "Card/Header"
	if strings.Contains(name, "/") || strings.Contains(name, ".") {
		normalized := strings.ReplaceAll(name, ".", "/")
		parts := strings.Split(normalized, "/")
		var pascalParts []string
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				pascalParts = append(pascalParts, toPascal(trimmed))
			}
		}
		if len(pascalParts) > 1 {
			parentSegments := pascalParts[:len(pascalParts)-1]
			basePascal := pascalParts[len(pascalParts)-1]
			dirParts := append([]string{root, "components"}, parentSegments...)
			dir := filepath.Join(dirParts...)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			target := filepath.Join(dir, basePascal+".kiw")
			if _, err := os.Stat(target); err == nil {
				rel, _ := filepath.Rel(root, target)
				return fmt.Errorf("component already exists at %s", filepath.ToSlash(rel))
			}
			pascal := strings.Join(pascalParts, ".")
			slug := toSlug(basePascal)
			content := componentTemplate(pascal, slug)
			if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, target)
			fmt.Printf("created %s\n", filepath.ToSlash(rel))
			return nil
		}
	}

	pascal := toPascal(name)
	slug := toSlug(name)
	dir := filepath.Join(root, "components")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	target := filepath.Join(dir, pascal+".kiw")
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("component %q already exists at components/%s.kiw", pascal, pascal)
	}

	content := componentTemplate(pascal, slug)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	fmt.Printf("created components/%s.kiw\n", pascal)
	return nil
}

func generatePage(root, name string) error {
	name = strings.TrimSuffix(name, ".kiw")
	slug := toSlug(name)
	title := toTitle(name)
	dir := filepath.Join(root, "pages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	target := filepath.Join(dir, slug+".kiw")
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("page %q already exists at pages/%s.kiw", slug, slug)
	}

	content := pageTemplate(title, slug)

	if err := safeWriteArtifact(root, target, []byte(content)); err != nil {
		return err
	}
	fmt.Printf("created pages/%s.kiw\n", slug)
	return nil
}
