package release

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/version"
)

func TestCompatibilityMatrix_InventoryAndOwnership(t *testing.T) {
	matrix := DefaultCompatibilityMatrix()
	if len(matrix) == 0 {
		t.Fatal("expected non-empty compatibility matrix")
	}

	seenIDs := make(map[string]bool)
	for _, b := range matrix {
		if b.ID == "" {
			t.Errorf("boundary missing ID: %+v", b)
		}
		if seenIDs[b.ID] {
			t.Errorf("duplicate boundary ID: %s", b.ID)
		}
		seenIDs[b.ID] = true

		if b.SourceModule == "" || b.TargetModule == "" {
			t.Errorf("boundary %s has empty source or target", b.ID)
		}
		if b.Owner == "" || !strings.HasPrefix(b.Owner, "@krewire/") {
			t.Errorf("boundary %s must have assigned owner starting with @krewire/, got %q", b.ID, b.Owner)
		}
		if b.Status != StatusSupported && b.Status != StatusBeta && b.Status != StatusExperimental && b.Status != StatusPlanned {
			t.Errorf("boundary %s has unknown status %q", b.ID, b.Status)
		}
	}
}

func TestRunQualification_Success(t *testing.T) {
	root, err := FindWorkspaceRoot()
	if err != nil {
		t.Skip("skipping workspace test outside go.work")
	}

	ctx := &QualificationContext{
		WorkspaceRoot: root,
		GitCommit:     "testcommit123",
	}

	report, err := RunQualification(ctx)
	if err != nil {
		t.Fatalf("RunQualification failed: %v", err)
	}

	if !report.OverallPassed {
		t.Errorf("expected qualification to pass, but OverallPassed=false; failedCount=%d", report.FailedCount)
		for _, r := range report.Results {
			if !r.Passed && !r.Skipped {
				t.Errorf("boundary %s failed: %s", r.BoundaryID, r.Error)
			}
		}
	}

	if report.PassedCount == 0 {
		t.Error("expected at least 1 passed boundary")
	}
	if report.SkippedCount == 0 {
		t.Error("expected skipped count > 0 for experimental/planned boundaries")
	}
	if report.GitCommit != "testcommit123" {
		t.Errorf("got GitCommit %q, want testcommit123", report.GitCommit)
	}

	// Verify formatting
	human := FormatHumanReadable(report)
	if !strings.Contains(human, "RELEASE QUALIFICATION REPORT") {
		t.Errorf("formatted text missing header: %s", human)
	}

	// Verify JSON export
	var buf bytes.Buffer
	if err := ExportJSON(report, &buf); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}
	var unmarshaled QualificationReport
	if err := json.Unmarshal(buf.Bytes(), &unmarshaled); err != nil {
		t.Fatalf("unmarshal exported JSON failed: %v", err)
	}
	if unmarshaled.GitCommit != "testcommit123" {
		t.Errorf("unmarshaled commit %q != testcommit123", unmarshaled.GitCommit)
	}
}

func TestRunQualification_VersionMismatchFails(t *testing.T) {
	ctx := &QualificationContext{
		WorkspaceRoot: "/tmp",
		GitCommit:     "sha123",
		ModuleVersions: map[string]version.Version{
			"krewire": version.MustParseVersion("0.2.0"),
			"ui":      version.MustParseVersion("0.4.0"), // Incompatible minor in 0.y.z!
		},
		ModuleRequirements: map[string]map[string]version.Version{},
	}

	report, err := RunQualification(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallPassed {
		t.Fatal("expected qualification to fail due to incompatible minor version between krewire and ui")
	}

	var foundFail bool
	for _, r := range report.Results {
		if r.BoundaryID == "krewire-ui" {
			if r.Passed {
				t.Error("expected krewire-ui boundary to fail")
			}
			if !strings.Contains(r.Error, "incompatible version pair") {
				t.Errorf("unexpected error message: %s", r.Error)
			}
			foundFail = true
		}
	}
	if !foundFail {
		t.Error("did not find krewire-ui in results")
	}
}

func TestRunQualification_ContractBreakageFails(t *testing.T) {
	// Custom boundary with failing contract
	ctx := &QualificationContext{
		WorkspaceRoot: t.TempDir(),
		GitCommit:     "dummy",
		ModuleVersions: map[string]version.Version{
			"krewire": version.MustParseVersion("0.2.0"),
			"ui":      version.MustParseVersion("0.2.0"),
		},
	}

	// Running in empty temp directory will fail krewire-ui because kiw.pkg.json is missing
	report, err := RunQualification(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallPassed {
		t.Fatal("expected qualification to fail in empty temp dir due to missing manifests")
	}

	if report.FailedCount == 0 {
		t.Fatal("expected failed count > 0")
	}
}

func TestRunQualification_ExperimentalSkippedDistinction(t *testing.T) {
	ctx := &QualificationContext{
		WorkspaceRoot: t.TempDir(),
		GitCommit:     "dummy",
		ModuleVersions: map[string]version.Version{
			"krewire": version.MustParseVersion("0.2.0"),
		},
		IncludeExperimental: false,
	}

	report, err := RunQualification(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var foundExperimental, foundPlanned bool
	for _, r := range report.Results {
		if r.BoundaryID == "cloud-mesh" {
			foundExperimental = true
			if !r.Skipped {
				t.Errorf("expected cloud-mesh to be skipped, got: %+v", r)
			}
			if !strings.Contains(r.Details, "Experimental") {
				t.Errorf("expected Experimental details, got: %s", r.Details)
			}
		}
		if r.BoundaryID == "dash-p2p" {
			foundPlanned = true
			if !r.Skipped {
				t.Errorf("expected dash-p2p to be skipped, got: %+v", r)
			}
			if !strings.Contains(r.Details, "Planned") {
				t.Errorf("expected Planned details, got: %s", r.Details)
			}
		}
	}
	if !foundExperimental {
		t.Error("cloud-mesh boundary not evaluated")
	}
	if !foundPlanned {
		t.Error("dash-p2p boundary not evaluated")
	}
}
