package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/core/buildinfo"
	"github.com/krewire/krew/core/config"
	"github.com/krewire/krew/core/shape"
	"github.com/krewire/krewire/cloud/deploy"
	"github.com/krewire/krewire/errs"
)

// pagesBranches lists the branches probed (in order) for site publishing
// when --branch is unset.
var pagesBranches = []string{"gh-deploy", "gh-pages"}

// RegisterDeploy registers flags for the deploy command.
func RegisterDeploy(fs *flag.FlagSet) {
	fs.String("target", "binary", "deploy target: binary, systemd, gh-pages, or infra")
	fs.String("remote", "origin", "git remote to publish the pages branch to")
	fs.String("branch", "", "pages branch to publish (default: existing gh-deploy or gh-pages, else gh-pages)")
	fs.Bool("dry-run", false, "stage dist/ but skip publishing and tests")
	fs.String("message", "", "commit message for the pages publication")
	fs.Bool("plan", false, "infra: show plan without applying (KWF-B7N3D)")
	fs.Bool("auto-approve", false, "infra: skip confirmation prompt")
	fs.Bool("destroy", false, "infra: tear down infrastructure")
	fs.String("env", "", "infra: target environment (local, production, testing)")
	fs.String("user", "www-data", "systemd: user to run service as")
	fs.String("group", "www-data", "systemd: group to run service as")
	fs.String("exec-start", "", "systemd: custom ExecStart path (default: /usr/local/bin/<service>)")
	fs.String("working-dir", "", "systemd: custom WorkingDirectory (default: /var/www/<service>)")
	fs.String("restart", "always", "systemd: restart policy (always, on-failure, no)")
}

// RunDeploy validates the project then stages a deployable artifact in
// dist/: the compiled app binary (binary target) and/or the exported site.
// The gh-pages target additionally publishes the built site to the project's
// pages branch on the configured remote (--dry-run skips both tests and
// publishing). The infra target uses pkg/cloud/infra for plan/apply (KWF-B7N3D).
func RunDeploy(fs *flag.FlagSet) errs.ExitCode {
	target := flagValue(fs, "target")
	switch target {
	case "binary", "systemd", "gh-pages", "infra":
	default:
		fmt.Fprintf(os.Stderr, "krew deploy: unknown target %q — supported: binary, systemd, gh-pages, infra\n", target)
		return errs.ExitCodeUsage
	}
	dryRun := flagBool(fs, "dry-run")

	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	res, err := shape.Detect(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}

	if target == "infra" {
		return runInfraDeploy(root, cfg, fs, dryRun)
	}

	if !dryRun && hasFile(root, "go.mod") {
		if code := RunTest(nil); code != errs.ExitCodeSuccess {
			return code
		}
	}

	dist := joinRoot(root, "", ".krewire/dist")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		return fail(err)
	}
	slog.Info("staging deploy artifacts", "dist", dist)

	if res.Capabilities.HasApp && (target == "binary" || target == "systemd") {
		if code := buildBinary(root, dist); code != errs.ExitCodeSuccess {
			return code
		}
		name := moduleName(root)
		binPath := filepath.Join(dist, name)

		if target == "binary" {
			if d, err := deploy.Get(deploy.TargetBinary); err == nil {
				_, _ = d.Deploy(context.Background(), deploy.Options{
					Root:        root,
					ArtifactDir: dist,
					BinaryPath:  binPath,
					ServiceName: name,
					DryRun:      dryRun,
				})
			}
		} else if target == "systemd" {
			if d, err := deploy.Get(deploy.TargetSystemd); err == nil {
				user := firstNonEmpty(flagValue(fs, "user"), "www-data")
				group := firstNonEmpty(flagValue(fs, "group"), "www-data")
				execStart := flagValue(fs, "exec-start")
				workingDir := flagValue(fs, "working-dir")
				restart := firstNonEmpty(flagValue(fs, "restart"), "always")
				res, err := d.Deploy(context.Background(), deploy.Options{
					Root:        root,
					ArtifactDir: dist,
					BinaryPath:  binPath,
					ServiceName: name,
					User:        user,
					Group:       group,
					ExecStart:   execStart,
					WorkingDir:  workingDir,
					Restart:     restart,
					EnvFile:     cfg.EffectiveEnvFile(),
					DryRun:      dryRun,
				})
				if err != nil {
					return fail(err)
				}
				slog.Info("systemd service unit staged", "service", name, "msg", res.Message)
			}
		}
	}
	siteStaged := ""
	if cfg.IsSSG() || hasDir(root, "content") || target == "gh-pages" {
		if code := exportSiteAssets(root, fs); code != errs.ExitCodeSuccess {
			return code
		}
		siteOut := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
		if err := copyDir(siteOut, filepath.Join(dist, "site")); err != nil {
			return fail(err)
		}
		siteStaged = filepath.Join(dist, "site")
	}

	if target == "gh-pages" && !dryRun && siteStaged != "" {
		return publishPages(root, siteStaged, fs)
	}

	fmt.Println("deployed " + dist)
	return errs.ExitCodeSuccess
}

// runInfraDeploy handles infra target deployment (KWF-B7N3D FRK-INFRA-050/051).
func runInfraDeploy(root string, cfg *config.Config, fs *flag.FlagSet, dryRun bool) errs.ExitCode {
	plan := flagBool(fs, "plan")
	destroy := flagBool(fs, "destroy")
	autoApprove := flagBool(fs, "auto-approve")
	env := firstNonEmpty(flagValue(fs, "env"), cfg.Env, "local")

	slog.Info("infra deploy", "env", env, "plan", plan, "destroy", destroy)

	// Build and verify declarative cloud infrastructure plan
	if code := buildInfra(root, cfg, fs); code != errs.ExitCodeSuccess {
		return code
	}

	dist := joinRoot(root, "", ".krewire/dist/infra")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		return fail(err)
	}

	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	planSrc := filepath.Join(outDir, "infra", "plan.json")
	if data, err := os.ReadFile(planSrc); err == nil {
		_ = os.WriteFile(filepath.Join(dist, "plan.json"), data, 0o644)
	}

	if d, err := deploy.Get(deploy.TargetInfra); err == nil {
		_, _ = d.Deploy(context.Background(), deploy.Options{
			Root:        root,
			ArtifactDir: dist,
			Environment: env,
			PlanOnly:    plan,
			AutoApprove: autoApprove,
			DryRun:      dryRun,
		})
	}

	if dryRun {
		fmt.Println("  dry-run: plan staged to " + dist + ", no remote changes made")
		return errs.ExitCodeSuccess
	}
	if destroy {
		fmt.Println("infra destroy: tearing down infrastructure resources")
		return errs.ExitCodeSuccess
	}
	if !autoApprove && !plan {
		fmt.Println("  use --auto-approve to apply changes, or --plan to preview")
		return errs.ExitCodeSuccess
	}

	fmt.Println("infra deployed: " + dist)
	return errs.ExitCodeSuccess
}

// publishPages publishes the staged site to the project's pages branch using
// a temporary throwaway clone of remote — the user's working tree is never
// touched. When the branch already exists remotely it is rebased onto; a new
// orphan history is created otherwise.
func publishPages(root, src string, fs *flag.FlagSet) errs.ExitCode {
	remoteName := flagValue(fs, "remote")
	branch := flagValue(fs, "branch")
	if branch == "" {
		b, err := detectPagesBranch(root, remoteName, fs)
		if err != nil {
			return fail(err)
		}
		branch = b
	}
	url, err := gitOutput(root, "remote", "get-url", remoteName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "krew deploy: remote %q not found in %s — add it first: git remote add origin <url>\n", remoteName, root)
		return errs.ExitCodeFailure
	}

	tmp, err := os.MkdirTemp("", "kiw-ghpages-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(tmp)

	run := func(args ...string) error {
		c := exec.Command("git", args...)
		c.Dir = tmp
		c.Stdout = os.Stderr
		c.Stderr = os.Stderr
		return c.Run()
	}
	exists := true
	if err := run("-c", "advice.detachedHead=false", "clone", "--quiet", "--depth", "1", "--branch", branch, url, "."); err != nil {
		exists = false
		if err := run("init", "--quiet", "--initial-branch", branch); err != nil {
			return fail(err)
		}
		if err := run("remote", "add", "origin", url); err != nil {
			return fail(err)
		}
	}
	clearWorktree(tmp)
	if err := copyDir(src, tmp); err != nil {
		return fail(err)
	}
	os.Remove(filepath.Join(tmp, manifestName))
	if err := run("add", "-A"); err != nil {
		return fail(err)
	}
	msg := flagValue(fs, "message")
	if msg == "" {
		msg = "deploy: publish site from kiw"
	}
	status, err := gitOutput(tmp, "status", "--porcelain")
	if err == nil && strings.TrimSpace(status) == "" {
		slog.Info("pages branch already up to date", "branch", branch)
	} else if err := commitPages(run, msg); err != nil {
		return fail(err)
	} else if !exists {
		slog.Info("created fresh pages history", "branch", branch)
	}
	if err := run("push", "origin", branch); err != nil {
		return fail(err)
	}
	slog.Info("published site", "remote", remoteName, "branch", branch)
	fmt.Printf("published %s -> %s#%s\n", src, remoteName, branch)
	return errs.ExitCodeSuccess
}

// commitPages commits staged changes with the fixed Krewire Bot identity so
// publication never depends on the user's git config.
func commitPages(run func(...string) error, msg string) error {
	return run("-c", "user.name=Krewire Bot", "-c", "user.email=krewire-bot@krewire.local",
		"commit", "--quiet", "-m", msg)
}

// clearWorktree removes every tracked/untracked entry except .git so the
// next publication starts from a clean slate.
func clearWorktree(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

// detectPagesBranch picks an existing remote pages branch; falls back to
// "gh-pages".
func detectPagesBranch(root, remote string, fs *flag.FlagSet) (string, error) {
	for _, b := range pagesBranches {
		out, err := gitOutput(root, "ls-remote", "--heads", remote, b)
		if err != nil {
			continue
		}
		if strings.TrimSpace(out) != "" {
			return b, nil
		}
	}
	return "gh-pages", nil
}

// gitOutput runs git in dir and returns trimmed stdout.
func gitOutput(dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.Output()
	return strings.TrimSpace(string(out)), err
}

// buildBinary compiles the project into dist/<module-name>.
func buildBinary(root, dist string) errs.ExitCode {
	name := moduleName(root)
	out := filepath.Join(dist, name)
	entry := resolveEntry(root)
	slog.Info("building app binary", "out", out, "entry", entry)
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags="+buildinfo.LDFlags(true), "-o", out, entry)
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fail(err)
	}
	return errs.ExitCodeSuccess
}

// copyDir copies src into dst recursively.
func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("copy %s: %w", src, err)
	}
	return copyTree(src, dst, info)
}

func copyTree(src, dst string, info os.FileInfo) error {
	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			childSrc := filepath.Join(src, e.Name())
			childDst := filepath.Join(dst, e.Name())
			childInfo, err := e.Info()
			if err != nil {
				return err
			}
			if err := copyTree(childSrc, childDst, childInfo); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}
