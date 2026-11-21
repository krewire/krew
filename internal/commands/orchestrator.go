package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/krewire/krew/internal/ecosystem"
	"github.com/krewire/krewire/pkg/errs"
)

// executeBuildDeps runs the list of dependency build commands sequentially
// in the project root directory, inheriting environment variables. If any
// command exits with a non-zero code, execution aborts and that exit code is returned.
// (FRK-BLD-001, FRK-BLD-002, FRK-BLD-003, FRK-BLD-004, FRK-BLD-005, FRK-BLD-010)
func executeBuildDeps(root string, deps []string, env []string) errs.ExitCode {
	for i, cmdStr := range deps {
		slog.Info("executing build dependency", "step", i+1, "total", len(deps), "command", cmdStr)
		fmt.Fprintf(os.Stdout, "── [%d/%d] %s ──\n", i+1, len(deps), cmdStr)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", cmdStr)
		} else {
			cmd = exec.Command("sh", "-c", cmdStr)
		}
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env

		if err := cmd.Run(); err != nil {
			slog.Error("build dependency failed", "command", cmdStr, "error", err)
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					return errs.ExitCode(status.ExitStatus())
				}
			}
			return errs.ExitCodeFailure
		}
	}
	return errs.ExitCodeSuccess
}

// RuntimeSuggestion represents an advisory build suggestion detected from project files.
type RuntimeSuggestion struct {
	Source  string // e.g. "package.json" or "composer.json"
	Command string // e.g. "npm run build" or "composer install"
	Hint    string // Human-readable rationale
}

// detectRuntimeBuildSuggestions inspects the project root for other runtime manifests
// (package.json, composer.json) and suggests build commands for build.depends_on.
// Detected commands are NOT auto-executed (FRK-BLD-020, FRK-BLD-021, FRK-BLD-022).
func detectRuntimeBuildSuggestions(root string) []RuntimeSuggestion {
	var suggestions []RuntimeSuggestion

	hosts := ecosystem.DetectAllHosts(root)
	for _, host := range hosts {
		if host.Type == ecosystem.HostNodeJS {
			pkgJSONPath := filepath.Join(root, ecosystem.ManifestPackageJSON)
			if data, err := os.ReadFile(pkgJSONPath); err == nil {
				var pkg struct {
					Scripts map[string]string `json:"scripts"`
				}
				if err := json.Unmarshal(data, &pkg); err == nil {
					if _, hasBuild := pkg.Scripts["build"]; hasBuild {
						cmd := host.PackageManager + " run build"
						if host.PackageManager == "yarn" {
							cmd = "yarn build"
						}
						suggestions = append(suggestions, RuntimeSuggestion{
							Source:  ecosystem.ManifestPackageJSON,
							Command: cmd,
							Hint:    "Node.js build script detected",
						})
					}
				}
			}
		} else if host.Type == ecosystem.HostLaravel || host.Type == ecosystem.HostPHP {
			suggestions = append(suggestions, RuntimeSuggestion{
				Source:  ecosystem.ManifestComposerJSON,
				Command: "composer install",
				Hint:    "PHP/Composer manifest detected",
			})
		}
	}

	return suggestions
}

// virtualRootFS creates an HTTP handler that serves files from multiple directory
// roots in order with extensionless resolution (FRK-BLD-032).
func virtualRootFS(roots ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := path.Clean("/" + r.URL.Path)
		if reqPath == "/" {
			reqPath = "/index.html"
		}

		for _, root := range roots {
			if root == "" {
				continue
			}
			fs := http.Dir(root)
			for _, candidate := range []string{reqPath, reqPath + ".html"} {
				f, err := fs.Open(candidate)
				if err != nil {
					continue
				}
				st, statErr := f.Stat()
				closeErr := f.Close()
				if statErr != nil || closeErr != nil || st.IsDir() {
					continue
				}
				http.ServeFile(w, r, filepath.Join(root, filepath.FromSlash(candidate)))
				return
			}
		}

		http.NotFound(w, r)
	})
}
