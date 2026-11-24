package krew_test

import (
	"bytes"
	"crypto/sha256"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func repoDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dir) == "conformance" {
		return filepath.Dir(dir)
	}
	return dir
}

// Verify pure Go execution without CGO dependencies.
func TestPlatform_PureGo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping pure Go check in short mode")
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(t.TempDir(), "krew"), "./cli")
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
			cmd := exec.Command("go", "build", "-o", outBin, "./cli")
			cmd.Dir = repoDir(t)
			cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0", "GOOS="+tgt.goos, "GOARCH="+tgt.goarch)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("cross-compile krew for %s/%s failed: %v\nOutput: %s", tgt.goos, tgt.goarch, err, string(out))
			}
		})
	}
}

// Automated verification that consecutive clean builds produce bitwise identical binaries.
func TestPlatform_BitwiseReproducibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bitwise reproducibility test in short mode")
	}
	tmpDir := t.TempDir()
	bin1 := filepath.Join(tmpDir, "krew-1")
	bin2 := filepath.Join(tmpDir, "krew-2")

	build := func(out string) []byte {
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", out, "./cli")
		cmd.Dir = repoDir(t)
		cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0")
		if outLog, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build failed: %v\nOutput: %s", err, string(outLog))
		}
		data, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("read binary: %v", err)
		}
		sum := sha256.Sum256(data)
		return sum[:]
	}

	h1 := build(bin1)
	h2 := build(bin2)
	if !bytes.Equal(h1, h2) {
		t.Fatalf("builds are not bitwise reproducible: %x != %x", h1, h2)
	}
}
