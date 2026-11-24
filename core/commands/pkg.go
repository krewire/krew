package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/model"
	kiwdsl "github.com/krewire/ui/kiw"
)

// RegisterPkg registers command-line flags for the pkg command.
func RegisterPkg(fs *flag.FlagSet) {
	fs.String("out", "", "pack: custom output archive path (.kpkg)")
	fs.Bool("json", false, "list: output installed packages in JSON format")
}

const pkgUsage = `usage:
  krew pkg init <name>               scaffold a new standalone component package
  krew pkg pack [dir]                bundle package into a distributable .kpkg archive
  krew pkg add <source>              install package from local dir or .kpkg archive
  krew pkg list                      list installed component packages in ./.krewire/pkg
  krew pkg health [name]             assess package health, stability tier, and compatibility
  krew pkg export <comp|dir>         export standalone component or package`

// RunPkg dispatches standalone component package commands.
func RunPkg(fs *flag.FlagSet) errs.ExitCode {
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, pkgUsage)
		return errs.ExitCodeUsage
	}

	root, err := findRoot()
	if err != nil {
		root, _ = os.Getwd()
	}

	switch args[0] {
	case "init":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: krew pkg init <name>")
			return errs.ExitCodeUsage
		}
		return pkgInit(root, args[1])

	case "pack":
		targetDir := root
		if len(args) > 1 {
			targetDir = joinRoot(root, args[1], "")
		}
		outPath := flagValue(fs, "out")
		return pkgPack(targetDir, outPath)

	case "add", "install":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: krew pkg add <source>")
			return errs.ExitCodeUsage
		}
		return pkgAdd(root, args[1])

	case "list", "ls":
		return pkgList(root, flagBool(fs, "json"))

	case "health", "status":
		pkgName := ""
		if len(args) > 1 {
			pkgName = args[1]
		}
		return pkgHealth(pkgName, flagBool(fs, "json"))

	case "export":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: krew pkg export <comp.kiw|dir>")
			return errs.ExitCodeUsage
		}
		return pkgExport(root, args[1])

	default:
		fmt.Fprintf(os.Stderr, "krew pkg: unknown subcommand %q\n%s\n", args[0], pkgUsage)
		return errs.ExitCodeUsage
	}
}

func pkgInit(root, name string) errs.ExitCode {
	targetDir := filepath.Join(root, name)
	if name == "." {
		targetDir = root
	}

	compsDir := filepath.Join(targetDir, "components")
	stylesDir := filepath.Join(targetDir, "styles")

	if err := os.MkdirAll(compsDir, 0o755); err != nil {
		return fail(err)
	}
	if err := os.MkdirAll(stylesDir, 0o755); err != nil {
		return fail(err)
	}

	cleanPkgName := name
	if !strings.HasPrefix(cleanPkgName, "@") && !strings.Contains(cleanPkgName, "/") {
		cleanPkgName = "@components/" + cleanPkgName
	}

	manifest := &kiwdsl.PackageManifest{
		Schema:      "https://krewire.com/schemas/kiw.pkg.schema.json",
		Name:        cleanPkgName,
		Version:     "0.1.0",
		Description: fmt.Sprintf("Independently distributable .kiw component package for %s", name),
		License:     "MIT",
		Main:        "./components",
		Components: map[string]string{
			"Button": "./components/Button.kiw",
		},
		Styles: "./styles/theme.css",
	}

	manifestPath := filepath.Join(targetDir, kiwdsl.PackageManifestFile)
	if err := kiwdsl.SavePackageManifest(manifestPath, manifest); err != nil {
		return fail(err)
	}

	// Scaffold starter Button.kiw
	buttonSrc := `<button class="btn" class:btn-primary={variant == 'primary'} class:btn-sm={size == 'sm'} {disabled}>
  <slot>Click me</slot>
</button>

<script>
  export let variant: string = "primary";
  export let size: string = "md";
  export let disabled: boolean = false;
</script>

<style scoped>
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 500;
  border-radius: 0.375rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
}
.btn-primary {
  background-color: #2563eb;
  color: #ffffff;
}
</style>
`
	if err := os.WriteFile(filepath.Join(compsDir, "Button.kiw"), []byte(buttonSrc), 0o644); err != nil {
		return fail(err)
	}

	cssSrc := `/* Standalone package theme */
:root {
  --pkg-primary: #2563eb;
}
`
	if err := os.WriteFile(filepath.Join(stylesDir, "theme.css"), []byte(cssSrc), 0o644); err != nil {
		return fail(err)
	}

	readme := fmt.Sprintf("# %s\n\nIndependently distributable `.kiw` component library.\n\n## Usage\n\n```html\n<script>\n  import Button from '%s/Button';\n</script>\n\n<Button variant=\"primary\">Hello World</Button>\n```\n", cleanPkgName, cleanPkgName)
	_ = os.WriteFile(filepath.Join(targetDir, "README.md"), []byte(readme), 0o644)

	fmt.Printf("✓ initialized standalone .kiw component package %s at %s\n", cleanPkgName, targetDir)
	fmt.Printf("  • Manifest:   %s\n", filepath.Join(targetDir, kiwdsl.PackageManifestFile))
	fmt.Printf("  • Component:  %s\n", filepath.Join(compsDir, "Button.kiw"))
	fmt.Printf("  • Stylesheet: %s\n", filepath.Join(stylesDir, "theme.css"))
	fmt.Printf("hint: run 'krew pkg pack' inside this directory to create a distributable %s archive\n", kiwdsl.PackageExtension)
	return errs.ExitCodeSuccess
}

func pkgPack(dir, outPath string) errs.ExitCode {
	archivePath, sha, err := kiwdsl.PackPackage(dir, outPath)
	if err != nil {
		return fail(err)
	}
	manifest, _ := kiwdsl.LoadPackageManifest(filepath.Join(dir, kiwdsl.PackageManifestFile))
	fmt.Printf("✓ packed %s@%s\n", manifest.Name, manifest.Version)
	fmt.Printf("  • Archive: %s\n", archivePath)
	fmt.Printf("  • SHA-256: %s\n", sha)
	fmt.Println("hint: distribute this archive or install with 'krew pkg add <archive>'")
	return errs.ExitCodeSuccess
}

func pkgAdd(root, source string) errs.ExitCode {
	modulesDir := filepath.Join(root, kiwdsl.DefaultPackageDir)
	if err := os.MkdirAll(modulesDir, 0o755); err != nil {
		return fail(err)
	}

	sourcePath := source
	if !filepath.IsAbs(sourcePath) {
		sourcePath = filepath.Join(root, source)
	}

	st, err := os.Stat(sourcePath)
	if err != nil {
		return fail(fmt.Errorf("source not found: %s", source))
	}

	var manifest *kiwdsl.PackageManifest

	if !st.IsDir() && strings.HasSuffix(sourcePath, kiwdsl.PackageExtension) {
		// Temporary unpack to inspect manifest
		tempDir, err := os.MkdirTemp("", "kiw-pkg-add-*")
		if err != nil {
			return fail(err)
		}
		defer os.RemoveAll(tempDir)

		tempManifest, err := kiwdsl.UnpackPackage(sourcePath, tempDir)
		if err != nil {
			return fail(err)
		}
		pkgDirName := strings.ReplaceAll(strings.TrimPrefix(tempManifest.Name, "@"), "/", "-")
		destDir := filepath.Join(modulesDir, pkgDirName)
		_ = os.RemoveAll(destDir)

		manifest, err = kiwdsl.UnpackPackage(sourcePath, destDir)
		if err != nil {
			return fail(err)
		}
	} else if st.IsDir() {
		// Directory containing kiw.pkg.json
		manifestPath := filepath.Join(sourcePath, kiwdsl.PackageManifestFile)
		m, err := kiwdsl.LoadPackageManifest(manifestPath)
		if err != nil {
			return fail(fmt.Errorf("%s is not a valid component package: %w", sourcePath, err))
		}
		manifest = m
		pkgDirName := strings.ReplaceAll(strings.TrimPrefix(manifest.Name, "@"), "/", "-")
		destDir := filepath.Join(modulesDir, pkgDirName)
		_ = os.RemoveAll(destDir)

		if err := copyDir(sourcePath, destDir); err != nil {
			return fail(err)
		}
	} else {
		return fail(fmt.Errorf("unsupported package source %q; expected %s archive or directory with kiw.pkg.json", source, kiwdsl.PackageExtension))
	}

	var compList []string
	for c := range manifest.Components {
		compList = append(compList, c)
	}
	sort.Strings(compList)

	fmt.Printf("✓ installed component package %s@%s into ./%s/%s\n",
		manifest.Name,
		manifest.Version,
		kiwdsl.DefaultPackageDir,
		strings.ReplaceAll(strings.TrimPrefix(manifest.Name, "@"), "/", "-"),
	)
	fmt.Printf("  • Components: %s\n", strings.Join(compList, ", "))
	if manifest.Styles != "" {
		fmt.Printf("  • Stylesheet: %s (auto-declared)\n", manifest.Styles)
	}
	return errs.ExitCodeSuccess
}

func pkgList(root string, printJSON bool) errs.ExitCode {
	modulesDir := filepath.Join(root, kiwdsl.DefaultPackageDir)
	var packages []*kiwdsl.PackageManifest

	scanDir := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			pManifest := filepath.Join(dir, e.Name(), kiwdsl.PackageManifestFile)
			if m, err := kiwdsl.LoadPackageManifest(pManifest); err == nil {
				duplicate := false
				for _, existing := range packages {
					if existing.Name == m.Name {
						duplicate = true
						break
					}
				}
				if !duplicate {
					packages = append(packages, m)
				}
			}
		}
	}

	scanDir(modulesDir)
	// Fallback check for legacy kiw_modules directory if distinct
	if modulesDir != filepath.Join(root, "kiw_modules") {
		scanDir(filepath.Join(root, "kiw_modules"))
	}

	if printJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(packages)
		return errs.ExitCodeSuccess
	}

	if len(packages) == 0 {
		fmt.Printf("no standalone component packages installed in ./%s\n", kiwdsl.DefaultPackageDir)
		fmt.Printf("hint: install with 'krew pkg add <dir|%s>'\n", kiwdsl.PackageExtension)
		return errs.ExitCodeSuccess
	}

	fmt.Printf("Installed .kiw component packages (%d):\n\n", len(packages))
	for _, p := range packages {
		var comps []string
		for c := range p.Components {
			comps = append(comps, c)
		}
		sort.Strings(comps)
		fmt.Printf("  • %s @ %s\n", p.Name, p.Version)
		if p.Description != "" {
			fmt.Printf("    %s\n", p.Description)
		}
		fmt.Printf("    Exports: %s\n", strings.Join(comps, ", "))
		if p.Styles != "" {
			fmt.Printf("    Styles:  %s\n", p.Styles)
		}
		fmt.Println()
	}
	return errs.ExitCodeSuccess
}

func pkgExport(root, target string) errs.ExitCode {
	targetPath := target
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, target)
	}

	st, err := os.Stat(targetPath)
	if err != nil {
		return fail(err)
	}

	if !st.IsDir() && strings.HasSuffix(targetPath, ".kiw") {
		mod, err := kiwdsl.ParseKiwFile(targetPath)
		if err != nil {
			return fail(err)
		}
		baseName := strings.TrimSuffix(filepath.Base(targetPath), ".kiw")
		fmt.Printf("✓ verified component %q is standalone exportable\n", baseName)
		fmt.Printf("  • Props declared: %d\n", len(mod.Props))
		fmt.Printf("  • Slots defined:  %d\n", len(mod.Slots))
		fmt.Printf("  • Styles scoped:  %d blocks\n", len(mod.Styles))
		return errs.ExitCodeSuccess
	}

	return pkgPack(targetPath, "")
}

func pkgHealth(name string, jsonOutput bool) errs.ExitCode {
	if name != "" {
		pkg, found := GetPackageMetadata(name)
		if !found {
			fmt.Fprintf(os.Stderr, "krew pkg health: unknown package %q (run 'krew pkg health' without arguments to see all)\n", name)
			return errs.ExitCodeUsage
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(pkg)
			return errs.ExitCodeSuccess
		}

		fmt.Printf("\n── Package Health Assessment: %s ──────────────────────────\n", pkg.Name)
		fmt.Printf("  • Import Path:     %s\n", pkg.ImportPath)
		fmt.Printf("  • Layer:           Layer %d\n", pkg.Layer)
		fmt.Printf("  • Stability:       %s\n", formatStability(pkg.Stability))
		fmt.Printf("  • Purpose:         %s\n", pkg.Purpose)
		fmt.Printf("  • License:         %s\n", pkg.License)
		fmt.Printf("  • Min Go Version:  %s\n", pkg.MinGoVersion)

		if len(pkg.ThirdPartyDeps) == 0 {
			fmt.Printf("  • Dependencies:    0 external dependencies (100%% pure Go standard library)\n")
		} else {
			fmt.Printf("  • Dependencies:    %s\n", strings.Join(pkg.ThirdPartyDeps, ", "))
		}

		fmt.Printf("  • Breaking Policy: %s\n", pkg.BreakingChangePolicy)
		if pkg.MigrationGuide != "" {
			fmt.Printf("  • Migration Guide: %s\n", pkg.MigrationGuide)
		}
		fmt.Println()
		return errs.ExitCodeSuccess
	}

	catalog := CanonicalPackageCatalog()
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(catalog)
		return errs.ExitCodeSuccess
	}

	fmt.Printf("\n── Canonical First-Party Package Health Catalog ────────────\n")
	fmt.Printf("  Total Packages: %d\n\n", len(catalog))

	for _, pkg := range catalog {
		fmt.Printf("  %-12s [Layer %d] [%-12s] %s\n", pkg.Name, pkg.Layer, pkg.Stability, pkg.Purpose)
	}

	fmt.Printf("\nRun 'krew pkg health <name>' for detailed dependency, compatibility, and migration contracts.\n\n")
	return errs.ExitCodeSuccess
}

func formatStability(tier model.StabilityTier) string {
	switch tier {
	case model.StabilityStable:
		return "✓ STABLE (Production Ready · Strict Backward Compatibility)"
	case model.StabilityBeta:
		return "⚠ BETA (Feature Complete · Pre-1.0 Stabilization)"
	case model.StabilityExperimental:
		return "ℹ EXPERIMENTAL (Active Exploration · API May Evolve)"
	case model.StabilityDeprecated:
		return "✗ DEPRECATED (Slated for Removal · Migration Required)"
	default:
		return string(tier)
	}
}
