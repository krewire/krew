package krew_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// Verify pure Go execution without CGO dependencies.
func TestPlatform_PureGo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping pure Go check in short mode")
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(t.TempDir(), "krew"), "./cmd/krew")
	cmd.Dir = repoDir(t)
	cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build with CGO_ENABLED=0 failed: %v\nOutput: %s", err, string(out))
	}
}

// Cross-compilation build verification for krew CLI on Linux, macOS, and Windows.
func TestPlatform_CrossCompileCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cross-compilation in short mode")
	}
	targets := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	for _, tgt := range targets {
		t.Run(tgt.goos+"_"+tgt.goarch, func(t *testing.T) {
			outBin := filepath.Join(t.TempDir(), "krew-"+tgt.goos+"-"+tgt.goarch)
			cmd := exec.Command("go", "build", "-o", outBin, "./cmd/krew")
			cmd.Dir = repoDir(t)
			cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0", "GOOS="+tgt.goos, "GOARCH="+tgt.goarch)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("cross-compile krew for %s/%s failed: %v\nOutput: %s", tgt.goos, tgt.goarch, err, string(out))
			}
		})
	}
}
