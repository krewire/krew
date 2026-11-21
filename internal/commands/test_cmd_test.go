package commands

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krew/internal/gomod"
	"github.com/krewire/krewire/pkg/errs"
)

func TestIsPackageFilter_Valid(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"./web", true},
		{"github.com/krewire/krewire/pkg/web", true},
		{"web_test.go", true},
		{"TestKWF", false},
		{"M4P", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isPackageFilter(c.in); got != c.want {
			t.Errorf("isPackageFilter(%q)=%v want %v", c.in, got, c.want)
		}
	}
}

func TestIsKnownTestFlag_Valid(t *testing.T) {
	if !isKnownTestFlag("--filter") {
		t.Error("isKnownTestFlag --filter should be true")
	}
	if !isKnownTestFlag("-watch") {
		t.Error("isKnownTestFlag -watch should be true")
	}
	if isKnownTestFlag("--unknown") {
		t.Error("isKnownTestFlag --unknown should be false")
	}
}

func TestResolveModuleWalkingUpward(t *testing.T) {
	tmp := t.TempDir()
	goModContent := "module example.com/testwalk\n\ngo 1.27.1\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goModContent), 0o600); err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(tmp, "sub", "deep", "dir")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	mod, err := gomod.Find(nested)
	if err != nil {
		t.Fatalf("expected to resolve module from %s: %v", nested, err)
	}
	if mod.Path != "example.com/testwalk" {
		t.Fatalf("got module path %q, want %q", mod.Path, "example.com/testwalk")
	}
}

func TestOutsideGoModuleExitCode(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	code := RunTest(nil)
	if code != errs.ExitCodeUsage {
		t.Fatalf("got exit code %v, want ExitCodeUsage (2)", code)
	}
}

func TestRunTest_Success(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()

	goMod := "module example.com/testpass\n\ngo 1.27.1\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	testCode := `package testpass

import "testing"

func TestSimplePass(t *testing.T) {
	if 1+1 != 2 {
		t.Fatal("math broke")
	}
}
`
	if err := os.WriteFile(filepath.Join(tmp, "pass_test.go"), []byte(testCode), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	code := RunTest(nil)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("got exit code %v, want ExitCodeSuccess (0)", code)
	}
}

func TestRunTest_Failure(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()

	goMod := "module example.com/testfail\n\ngo 1.27.1\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	testCode := `package testfail

import "testing"

func TestSimpleFail(t *testing.T) {
	t.Fatal("deliberate failure for test")
}
`
	if err := os.WriteFile(filepath.Join(tmp, "fail_test.go"), []byte(testCode), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	code := RunTest(nil)
	if code != errs.ExitCodeFailure {
		t.Fatalf("got exit code %v, want ExitCodeFailure (1)", code)
	}
}

func TestIsKnownTestFlag_CoverageFlags(t *testing.T) {
	flags := []string{
		"-cover", "--cover",
		"-coverage", "--coverage",
		"-coverprofile", "--coverprofile",
		"-html", "--html",
		"-cover-html", "--cover-html",
		"-threshold", "--threshold",
		"-min", "--min",
		"-summary", "--summary",
		"-breakdown", "--breakdown",
	}
	for _, f := range flags {
		if !isKnownTestFlag(f) {
			t.Errorf("isKnownTestFlag(%q) = false, want true", f)
		}
	}
}

func TestResolveCoverageOptions(t *testing.T) {
	// Nil flagset
	opts, err := resolveCoverageOptions(nil)
	if err != nil || opts.Enabled {
		t.Fatalf("expected disabled coverage for nil flagset, got %+v, err: %v", opts, err)
	}

	// Default empty flagset
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	RegisterTest(fs)
	opts, err = resolveCoverageOptions(fs)
	if err != nil || opts.Enabled {
		t.Fatalf("expected disabled coverage by default, got %+v, err: %v", opts, err)
	}

	// Flagset with --cover
	fs = flag.NewFlagSet("test", flag.ContinueOnError)
	RegisterTest(fs)
	_ = fs.Parse([]string{"--cover"})
	opts, err = resolveCoverageOptions(fs)
	if err != nil {
		t.Fatalf("resolveCoverageOptions error: %v", err)
	}
	if !opts.Enabled || !opts.TempProfile {
		t.Fatalf("expected enabled with TempProfile, got %+v", opts)
	}
	_ = os.Remove(opts.ProfilePath)

	// Flagset with explicit profile and html
	fs = flag.NewFlagSet("test", flag.ContinueOnError)
	RegisterTest(fs)
	_ = fs.Parse([]string{"--coverprofile=custom.out", "--html=custom.html", "--threshold=85.5", "--summary"})
	opts, err = resolveCoverageOptions(fs)
	if err != nil {
		t.Fatalf("resolveCoverageOptions error: %v", err)
	}
	if !opts.Enabled || opts.TempProfile || opts.ProfilePath != "custom.out" || opts.HTMLPath != "custom.html" || opts.Threshold != 85.5 || !opts.Summary {
		t.Fatalf("unexpected coverage options: %+v", opts)
	}
}

func TestProcessCoverageResults_SummaryAndThreshold(t *testing.T) {
	tmp := t.TempDir()
	profilePath := filepath.Join(tmp, "cover.out")
	profileData := `mode: set
example.com/foo/foo.go:10.1,12.5 2 1
example.com/foo/foo.go:14.1,15.5 1 0
`
	if err := os.WriteFile(profilePath, []byte(profileData), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	opts := CoverageOptions{
		Enabled:      true,
		ProfilePath:  profilePath,
		Summary:      true,
		Threshold:    50.0,
		HasThreshold: true,
	}

	code := processCoverageResults(opts, &stdout, &stderr)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected success, got %v: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "66.7%") {
		t.Errorf("stdout missing coverage percentage: %s", stdout.String())
	}

	// Now fail threshold (need 80%, got 66.7%)
	stdout.Reset()
	stderr.Reset()
	opts.Threshold = 80.0
	code = processCoverageResults(opts, &stdout, &stderr)
	if code != errs.ExitCodeFailure {
		t.Fatalf("expected failure when below threshold, got %v", code)
	}
	if !strings.Contains(stderr.String(), "FAIL") {
		t.Errorf("stderr missing FAIL message: %s", stderr.String())
	}
}

func TestRunTest_WithCoverageAndSummary(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()

	goMod := "module example.com/testcov\n\ngo 1.27.1\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	libCode := `package testcov

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(filepath.Join(tmp, "lib.go"), []byte(libCode), 0o600); err != nil {
		t.Fatal(err)
	}

	testCode := `package testcov

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Fatal("addition failed")
	}
}
`
	if err := os.WriteFile(filepath.Join(tmp, "lib_test.go"), []byte(testCode), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	// Test 1: Run with --cover --summary
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	RegisterTest(fs)
	_ = fs.Parse([]string{"--cover", "--summary"})

	code := RunTest(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected RunTest with --cover --summary to pass, got %v", code)
	}

	// Test 2: Run with --threshold=100.0 (passes since 100% of Add is covered)
	fs = flag.NewFlagSet("test", flag.ContinueOnError)
	RegisterTest(fs)
	_ = fs.Parse([]string{"--threshold=100.0"})

	code = RunTest(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected RunTest with --threshold=100.0 to pass, got %v", code)
	}

	// Test 3: Run with --html report generation
	htmlOut := filepath.Join(tmp, "reports", "coverage.html")
	fs = flag.NewFlagSet("test", flag.ContinueOnError)
	RegisterTest(fs)
	_ = fs.Parse([]string{"--cover", "--html=" + htmlOut})

	code = RunTest(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected RunTest with --html to pass, got %v", code)
	}
	if _, err := os.Stat(htmlOut); os.IsNotExist(err) {
		t.Fatalf("expected HTML report file %s to be created", htmlOut)
	}

	// Test 4: Run with threshold failure (require 100%, but add uncovered function)
	uncoveredCode := `package testcov

func Add(a, b int) int {
	return a + b
}

func Uncovered(x int) int {
	if x > 10 {
		return x * 2
	}
	return x
}
`
	if err := os.WriteFile(filepath.Join(tmp, "lib.go"), []byte(uncoveredCode), 0o600); err != nil {
		t.Fatal(err)
	}

	fs = flag.NewFlagSet("test", flag.ContinueOnError)
	RegisterTest(fs)
	_ = fs.Parse([]string{"--threshold=100.0"})

	code = RunTest(fs)
	if code != errs.ExitCodeFailure {
		t.Fatalf("expected RunTest to fail when coverage is below threshold, got %v", code)
	}
}
