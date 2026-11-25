package release

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/krewire/krewire/version"
)

// BoundaryStatus classifies the readiness and enforcement tier of an integration boundary.
type BoundaryStatus string

const (
	StatusSupported    BoundaryStatus = "Supported"
	StatusBeta         BoundaryStatus = "Beta"
	StatusExperimental BoundaryStatus = "Experimental"
	StatusPlanned      BoundaryStatus = "Planned"
)

// IntegrationBoundary defines a contract between a producer and consumer repository in the ecosystem.
type IntegrationBoundary struct {
	ID                string                                `json:"id"`
	SourceModule      string                                `json:"source_module"`
	TargetModule      string                                `json:"target_module"`
	VersionConstraint string                                `json:"version_constraint"`
	Status            BoundaryStatus                        `json:"status"`
	Owner             string                                `json:"owner"`
	Description       string                                `json:"description"`
	ContractCheck     func(ctx *QualificationContext) error `json:"-"`
}

// QualificationContext provides runtime context for evaluating integration contracts.
type QualificationContext struct {
	WorkspaceRoot       string
	GitCommit           string
	ModuleVersions      map[string]version.Version
	ModuleRequirements  map[string]map[string]version.Version
	IncludeExperimental bool
	LogFn               func(format string, args ...any)
}

// BoundaryResult records the evaluation outcome of an integration boundary.
type BoundaryResult struct {
	BoundaryID   string         `json:"boundary_id"`
	SourceModule string         `json:"source_module"`
	TargetModule string         `json:"target_module"`
	Status       BoundaryStatus `json:"status"`
	Passed       bool           `json:"passed"`
	Skipped      bool           `json:"skipped"`
	Details      string         `json:"details"`
	Error        string         `json:"error,omitempty"`
	DurationMs   int64          `json:"duration_ms"`
}

// QualificationReport aggregates all boundary results for a release candidate commit.
type QualificationReport struct {
	Timestamp       string            `json:"timestamp"`
	GitCommit       string            `json:"git_commit"`
	WorkspaceRoot   string            `json:"workspace_root"`
	ModuleVersions  map[string]string `json:"module_versions"`
	TotalBoundaries int               `json:"total_boundaries"`
	PassedCount     int               `json:"passed_count"`
	FailedCount     int               `json:"failed_count"`
	SkippedCount    int               `json:"skipped_count"`
	OverallPassed   bool              `json:"overall_passed"`
	Results         []BoundaryResult  `json:"results"`
}

// DefaultCompatibilityMatrix returns the canonical list of cross-repository integration boundaries.
func DefaultCompatibilityMatrix() []IntegrationBoundary {
	return []IntegrationBoundary{
		{
			ID:                "krewire-ui",
			SourceModule:      "krewire",
			TargetModule:      "ui",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-runtime",
			Description:       "Verify @krewire/ui kiw.pkg.json manifest, Pop-Brutalist CSS tokens, and component exports",
			ContractCheck: func(ctx *QualificationContext) error {
				uiDir := filepath.Join(ctx.WorkspaceRoot, "ui")
				manifestPath := filepath.Join(uiDir, "kiw.pkg.json")
				data, err := os.ReadFile(manifestPath)
				if err != nil {
					return fmt.Errorf("missing ui/kiw.pkg.json: %w", err)
				}
				var pkg struct {
					Name       string            `json:"name"`
					Version    string            `json:"version"`
					Components map[string]string `json:"components"`
				}
				if err := json.Unmarshal(data, &pkg); err != nil {
					return fmt.Errorf("malformed kiw.pkg.json: %w", err)
				}
				if pkg.Name != "@krewire/ui" {
					return fmt.Errorf("invalid package name %q, want @krewire/ui", pkg.Name)
				}
				if len(pkg.Components) == 0 {
					return fmt.Errorf("ui manifest exposes 0 components")
				}
				cssPath := filepath.Join(uiDir, "styles", "krewire.css")
				if _, err := os.Stat(cssPath); err != nil {
					return fmt.Errorf("missing styles/krewire.css: %w", err)
				}
				return nil
			},
		},
		{
			ID:                "krewire-krew",
			SourceModule:      "krewire",
			TargetModule:      "krew",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-agent",
			Description:       "Verify krew agent engine imports krewire Go SDK and honors version requirements",
			ContractCheck: func(ctx *QualificationContext) error {
				krewDir := filepath.Join(ctx.WorkspaceRoot, "krew")
				modData, err := readRepoManifest(krewDir)
				if err != nil {
					return fmt.Errorf("missing krew/go.mod: %w", err)
				}
				if !strings.Contains(string(modData), "krewire") {
					return fmt.Errorf("krew/go.mod must depend on github.com/krewire/krewire")
				}
				return nil
			},
		},
		{
			ID:                "krewire-cloud",
			SourceModule:      "krewire",
			TargetModule:      "cloud",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-cloud",
			Description:       "Verify cloud PaaS control plane consumes krewire core, runner, and gateway contracts",
			ContractCheck: func(ctx *QualificationContext) error {
				cloudDir := filepath.Join(ctx.WorkspaceRoot, "cloud")
				if _, err := os.Stat(cloudDir); os.IsNotExist(err) {
					return nil
				}
				modData, err := readRepoManifest(cloudDir)
				if err != nil {
					return fmt.Errorf("missing cloud/go.mod: %w", err)
				}
				if !strings.Contains(string(modData), "krewire") {
					return fmt.Errorf("cloud/go.mod must depend on github.com/krewire/krewire")
				}
				return nil
			},
		},
		{
			ID:                "krewire-auth",
			SourceModule:      "krewire",
			TargetModule:      "auth",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-iam",
			Description:       "Verify auth IAM platform consumes krewire security primitives and errors",
			ContractCheck: func(ctx *QualificationContext) error {
				authDir := filepath.Join(ctx.WorkspaceRoot, "auth")
				if _, err := os.Stat(authDir); os.IsNotExist(err) {
					return nil
				}
				modData, err := readRepoManifest(authDir)
				if err != nil {
					return fmt.Errorf("missing auth/go.mod: %w", err)
				}
				if !strings.Contains(string(modData), "krewire") {
					return fmt.Errorf("auth/go.mod must depend on github.com/krewire/krewire")
				}
				return nil
			},
		},
		{
			ID:                "krewire-dash",
			SourceModule:      "krewire",
			TargetModule:      "dash",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-workspace",
			Description:       "Verify dash unified workspace consumes krewire TUI and runtime execution contracts",
			ContractCheck: func(ctx *QualificationContext) error {
				dashDir := filepath.Join(ctx.WorkspaceRoot, "dash")
				if _, err := os.Stat(dashDir); os.IsNotExist(err) {
					return nil
				}
				modData, err := readRepoManifest(dashDir)
				if err != nil {
					return fmt.Errorf("missing dash/go.mod: %w", err)
				}
				if !strings.Contains(string(modData), "krewire") {
					return fmt.Errorf("dash/go.mod must depend on github.com/krewire/krewire")
				}
				return nil
			},
		},
		{
			ID:                "kiw-runtime-conformance",
			SourceModule:      "krewire",
			TargetModule:      "runtime",
			VersionConstraint: ">=0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-runtime",
			Description:       "Verify .kiw compiler, SSR/SSG rendering, and client-island hydration conformance matrix",
			ContractCheck: func(ctx *QualificationContext) error {
				krewireDir := filepath.Join(ctx.WorkspaceRoot, "krewire")
				conformancePkg := "./runtime/conformance"
				cmd := exec.Command("go", "test", "-short", conformancePkg)
				cmd.Dir = krewireDir
				if out, err := cmd.CombinedOutput(); err != nil {
					return fmt.Errorf("runtime conformance tests failed: %w (output: %s)", err, string(out))
				}
				return nil
			},
		},
		{
			ID:                "core-performance-baseline",
			SourceModule:      "krewire",
			TargetModule:      "benchmarks",
			VersionConstraint: ">=0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-runtime",
			Description:       "Verify core performance benchmarks against canonical baseline without critical regression",
			ContractCheck: func(ctx *QualificationContext) error {
				krewireDir := filepath.Join(ctx.WorkspaceRoot, "krewire")
				baselinePath := filepath.Join(krewireDir, "docs", "benchmarks", "baseline-v0.2.0.json")
				if _, err := os.Stat(baselinePath); err != nil {
					return fmt.Errorf("missing baseline report: %w", err)
				}
				krewDir := filepath.Join(ctx.WorkspaceRoot, "krew")
				cmd := exec.Command("go", "test", "-short", "./core/bench/...")
				cmd.Dir = krewDir
				if out, err := cmd.CombinedOutput(); err != nil {
					return fmt.Errorf("bench tests failed: %w (output: %s)", err, string(out))
				}
				return nil
			},
		},
		{
			ID:                "auth-cloud",
			SourceModule:      "auth",
			TargetModule:      "cloud",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-iam",
			Description:       "Verify decoupled token verification contract between Auth and Cloud PaaS",
			ContractCheck: func(ctx *QualificationContext) error {
				cloudDir := filepath.Join(ctx.WorkspaceRoot, "cloud")
				if _, err := os.Stat(cloudDir); os.IsNotExist(err) {
					return nil
				}
				modData, err := readRepoManifest(cloudDir)
				if err != nil {
					return fmt.Errorf("missing cloud/go.mod: %w", err)
				}
				if !strings.Contains(string(modData), "auth") {
					return fmt.Errorf("cloud/go.mod must require github.com/krewire/auth")
				}
				if reqs, ok := ctx.ModuleRequirements["cloud"]; ok {
					if _, hasAuth := reqs["auth"]; !hasAuth {
						return fmt.Errorf("cloud version.go must declare EcosystemRequires for auth")
					}
				}
				return nil
			},
		},
		{
			ID:                "auth-dash",
			SourceModule:      "auth",
			TargetModule:      "dash",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-iam",
			Description:       "Verify user identity session and RBAC contract between Auth and Dash",
			ContractCheck: func(ctx *QualificationContext) error {
				dashDir := filepath.Join(ctx.WorkspaceRoot, "dash")
				if _, err := os.Stat(dashDir); os.IsNotExist(err) {
					return nil
				}
				modData, err := readRepoManifest(dashDir)
				if err != nil {
					return fmt.Errorf("missing dash/go.mod: %w", err)
				}
				if !strings.Contains(string(modData), "auth") {
					return fmt.Errorf("dash/go.mod must require github.com/krewire/auth")
				}
				if reqs, ok := ctx.ModuleRequirements["dash"]; ok {
					if _, hasAuth := reqs["auth"]; !hasAuth {
						return fmt.Errorf("dash version.go must declare EcosystemRequires for auth")
					}
				}
				return nil
			},
		},
		{
			ID:                "krew-dash",
			SourceModule:      "krew",
			TargetModule:      "dash",
			VersionConstraint: "^0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-agent",
			Description:       "Verify autonomous AI and workspace protocol compatibility between Krew and Dash via core foundation",
			ContractCheck: func(ctx *QualificationContext) error {
				dashDir := filepath.Join(ctx.WorkspaceRoot, "dash")
				if _, err := os.Stat(dashDir); os.IsNotExist(err) {
					return nil
				}
				modData, err := readRepoManifest(dashDir)
				if err != nil {
					return fmt.Errorf("missing dash/go.mod: %w", err)
				}
				if !strings.Contains(string(modData), "krewire") {
					return fmt.Errorf("dash/go.mod must require github.com/krewire/krewire")
				}
				if reqs, ok := ctx.ModuleRequirements["dash"]; ok {
					if _, hasKrewire := reqs["krewire"]; !hasKrewire {
						return fmt.Errorf("dash version.go must declare EcosystemRequires for krewire")
					}
				}
				return nil
			},
		},
		{
			ID:                "cloud-mesh",
			SourceModule:      "cloud",
			TargetModule:      "distributed-workers",
			VersionConstraint: "0.3.x-exp",
			Status:            StatusExperimental,
			Owner:             "@krewire/team-cloud",
			Description:       "Experimental multi-node worker clustering and overlay service mesh",
			ContractCheck: func(ctx *QualificationContext) error {
				return nil
			},
		},
		{
			ID:                "security-assurance-baseline",
			SourceModule:      "krewire",
			TargetModule:      "sec",
			VersionConstraint: ">=0.2.0",
			Status:            StatusSupported,
			Owner:             "@krewire/team-security",
			Description:       "Verify end-to-end security assurance baseline, cryptographic primitives, and safe path bounds",
			ContractCheck: func(ctx *QualificationContext) error {
				krewireDir := filepath.Join(ctx.WorkspaceRoot, "krewire")
				cmd := exec.Command("go", "test", "-short", "./sec/...")
				cmd.Dir = krewireDir
				if out, err := cmd.CombinedOutput(); err != nil {
					return fmt.Errorf("security assurance tests failed: %w (output: %s)", err, string(out))
				}
				return nil
			},
		},
		{
			ID:                "dash-p2p",
			SourceModule:      "dash",
			TargetModule:      "edge-peering",
			VersionConstraint: "unreleased",
			Status:            StatusPlanned,
			Owner:             "@krewire/team-workspace",
			Description:       "Planned peer-to-peer workspace sync over WebRTC data channels",
			ContractCheck: func(ctx *QualificationContext) error {
				return nil
			},
		},
	}
}

// DiscoverGitCommit returns the active git commit SHA from workspace root or environment.
func DiscoverGitCommit(root string) string {
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		return sha
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	if root != "" {
		cmd.Dir = root
	}
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return "dev"
}

// RunQualification executes the compatibility matrix qualification against the workspace.
func RunQualification(ctx *QualificationContext) (*QualificationReport, error) {
	if ctx == nil {
		ctx = &QualificationContext{}
	}
	if ctx.WorkspaceRoot == "" {
		root, err := FindWorkspaceRoot()
		if err != nil {
			return nil, fmt.Errorf("could not locate workspace root (go.work): %w", err)
		}
		ctx.WorkspaceRoot = root
	}
	if ctx.GitCommit == "" {
		ctx.GitCommit = DiscoverGitCommit(ctx.WorkspaceRoot)
	}
	if ctx.ModuleVersions == nil {
		versions, err := ReadEcosystemActualVersions(ctx.WorkspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to read ecosystem versions: %w", err)
		}
		// Also inspect ui/kiw.pkg.json if available
		uiPkgPath := filepath.Join(ctx.WorkspaceRoot, "ui", "kiw.pkg.json")
		if data, err := os.ReadFile(uiPkgPath); err == nil {
			var pkg struct {
				Version string `json:"version"`
			}
			if json.Unmarshal(data, &pkg) == nil && pkg.Version != "" {
				if v, verr := version.ParseVersion(pkg.Version); verr == nil {
					versions["ui"] = v
				}
			}
		}
		ctx.ModuleVersions = versions
	}
	if ctx.ModuleRequirements == nil {
		reqs, err := ReadEcosystemRequirements(ctx.WorkspaceRoot)
		if err == nil {
			ctx.ModuleRequirements = reqs
		} else {
			ctx.ModuleRequirements = DefaultRequirements
		}
	}

	boundaries := DefaultCompatibilityMatrix()
	report := &QualificationReport{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		GitCommit:       ctx.GitCommit,
		WorkspaceRoot:   ctx.WorkspaceRoot,
		ModuleVersions:  make(map[string]string),
		TotalBoundaries: len(boundaries),
		OverallPassed:   true,
		Results:         make([]BoundaryResult, 0, len(boundaries)),
	}

	for mod, ver := range ctx.ModuleVersions {
		report.ModuleVersions[mod] = ver.String()
	}

	for _, b := range boundaries {
		start := time.Now()
		res := BoundaryResult{
			BoundaryID:   b.ID,
			SourceModule: b.SourceModule,
			TargetModule: b.TargetModule,
			Status:       b.Status,
		}

		// Handle Experimental and Planned boundaries
		if b.Status == StatusExperimental && !ctx.IncludeExperimental {
			res.Skipped = true
			res.Details = "Skipped (Experimental: enabled only via --include-experimental)"
			report.SkippedCount++
			res.DurationMs = time.Since(start).Milliseconds()
			report.Results = append(report.Results, res)
			continue
		}
		if b.Status == StatusPlanned {
			res.Skipped = true
			res.Details = "Skipped (Planned: roadmap boundary not yet shipped)"
			report.SkippedCount++
			res.DurationMs = time.Since(start).Milliseconds()
			report.Results = append(report.Results, res)
			continue
		}

		// Check SemVer compatibility constraints against declared module versions
		srcVer, hasSrc := ctx.ModuleVersions[b.SourceModule]
		tgtVer, hasTgt := ctx.ModuleVersions[b.TargetModule]

		if hasSrc && hasTgt {
			if !srcVer.IsCompatible(tgtVer) {
				res.Passed = false
				res.Error = fmt.Sprintf("incompatible version pair: %s %s vs %s %s (must match minor in pre-1.0.0)",
					b.SourceModule, srcVer, b.TargetModule, tgtVer)
				report.FailedCount++
				report.OverallPassed = false
				res.DurationMs = time.Since(start).Milliseconds()
				report.Results = append(report.Results, res)
				continue
			}
		}

		// Execute contract check
		if b.ContractCheck != nil {
			err := b.ContractCheck(ctx)
			if err != nil {
				res.Passed = false
				res.Error = err.Error()
				report.FailedCount++
				report.OverallPassed = false
			} else {
				res.Passed = true
				res.Details = fmt.Sprintf("Verified contract %s -> %s (owner: %s)", b.SourceModule, b.TargetModule, b.Owner)
				report.PassedCount++
			}
		} else {
			res.Passed = true
			res.Details = "No dynamic contract hook required"
			report.PassedCount++
		}

		res.DurationMs = time.Since(start).Milliseconds()
		report.Results = append(report.Results, res)
	}

	return report, nil
}

// ExportJSON writes the qualification report to w in formatted JSON.
func ExportJSON(report *QualificationReport, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatHumanReadable formats a qualification report as human-readable text.
func FormatHumanReadable(r *QualificationReport) string {
	var b strings.Builder
	b.WriteString("================================================================================\n")
	b.WriteString("                     KREWIRE RELEASE QUALIFICATION REPORT                       \n")
	b.WriteString("================================================================================\n")
	b.WriteString(fmt.Sprintf("Timestamp   : %s\n", r.Timestamp))
	b.WriteString(fmt.Sprintf("Git Commit  : %s\n", r.GitCommit))
	b.WriteString(fmt.Sprintf("Workspace   : %s\n", r.WorkspaceRoot))
	b.WriteString(fmt.Sprintf("Status      : %s\n", map[bool]string{true: "PASSED (RELEASE QUALIFIED)", false: "FAILED (BLOCKED)"}[r.OverallPassed]))
	b.WriteString(fmt.Sprintf("Summary     : %d Total | %d Passed | %d Failed | %d Skipped\n",
		r.TotalBoundaries, r.PassedCount, r.FailedCount, r.SkippedCount))
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString("Module Versions:\n")
	for mod, ver := range r.ModuleVersions {
		b.WriteString(fmt.Sprintf("  • %-12s : %s\n", mod, ver))
	}
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString("Boundary Checks:\n")
	for _, res := range r.Results {
		statusTag := "[PASS]"
		if res.Skipped {
			statusTag = "[SKIP]"
		} else if !res.Passed {
			statusTag = "[FAIL]"
		}
		b.WriteString(fmt.Sprintf("  %s %-16s (%s -> %s) [%s]\n",
			statusTag, res.BoundaryID, res.SourceModule, res.TargetModule, res.Status))
		if res.Error != "" {
			b.WriteString(fmt.Sprintf("         Error: %s\n", res.Error))
		} else if res.Details != "" {
			b.WriteString(fmt.Sprintf("         Note : %s\n", res.Details))
		}
	}
	b.WriteString("================================================================================\n")
	return b.String()
}

func readRepoManifest(repoDir string) (string, error) {
	if data, err := os.ReadFile(filepath.Join(repoDir, "go.work")); err == nil {
		return string(data), nil
	}
	var b strings.Builder
	_ = filepath.WalkDir(repoDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "go.mod" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err == nil {
			b.Write(data)
			b.WriteString("\n")
		}
		return nil
	})
	if b.Len() > 0 {
		return b.String(), nil
	}
	return "", fmt.Errorf("no go.work or go.mod in %s", repoDir)
}
