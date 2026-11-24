package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krewire/sec"
)

// ErrUnknownKind is returned when an unsupported artifact kind is requested.
var ErrUnknownKind = errors.New("unknown kind")

// MakeArtifact scaffolds a new component, page, handler, migration, model,
// middleware, or service in the project root.
func MakeArtifact(root, kind, name string) error {
	switch strings.ToLower(kind) {
	case "component":
		return generateComponent(root, name)
	case "page":
		return generatePage(root, name)
	case "handler":
		return generateHandler(root, name)
	case "migration":
		return generateMigration(root, name)
	case "model":
		return generateModel(root, name)
	case "seeder", "seed":
		return generateSeeder(root, name)
	case "factory":
		return generateFactory(root, name)
	case "repository", "repo":
		return generateRepository(root, name)
	case "middleware":
		return generateMiddleware(root, name)
	case "service":
		return generateService(root, name)
	case "job", "worker":
		return generateJob(root, name)
	case "chapter", "doc", "docs":
		return generateChapter(root, name)
	default:
		return fmt.Errorf("%w %q — supported: component, page, handler, migration, model, seeder, factory, repository, middleware, service, job, chapter, doc", ErrUnknownKind, kind)
	}
}

func hasDir(root, path string) bool {
	info, err := os.Stat(filepath.Join(root, path))
	return err == nil && info.IsDir()
}

// safeWriteArtifact validates that target resides strictly within root using sec.SafePath before writing.
func safeWriteArtifact(root, target string, content []byte) error {
	safe, err := sec.SafePath(root, target)
	if err != nil {
		return err
	}
	return os.WriteFile(safe, content, 0o644)
}
