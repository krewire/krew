package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krewire/errs"
)

// RegisterConfig registers flags for the config command.
func RegisterConfig(fs *flag.FlagSet) {
	fs.Bool("strict", false, "treat warnings (unknown or deprecated keys) as failures")
	fs.Bool("json", false, "print validation issues as JSON")
}

// RunConfig validates krewire.yaml against the JSON Schema or prints the schema.
//
//	krew config [validate] [path]   validate ./krewire.yaml (or path)
//	krew config schema              print the JSON Schema (Draft-07)
func RunConfig(fs *flag.FlagSet) errs.ExitCode {
	args := fs.Args()
	sub := "validate"
	if len(args) > 0 && (args[0] == "validate" || args[0] == "schema") {
		sub, args = args[0], args[1:]
	}
	if sub == "schema" {
		_, _ = os.Stdout.Write(config.SchemaJSON)
		return errs.ExitCodeSuccess
	}
	path := "krewire.yaml"
	if len(args) > 0 {
		path = args[0]
		if st, err := os.Stat(path); err == nil && st.IsDir() {
			path = filepath.Join(path, "krewire.yaml")
		}
	}
	issues, err := config.ValidateFile(path)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "krew config: %s not found — run 'krew init' to create one\n", path)
		return errs.ExitCodeUsage
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "krew config: %v\n", err)
		return errs.ExitCodeFailure
	}
	if flagBool(fs, "json") {
		out, _ := json.MarshalIndent(map[string]any{"file": path, "issues": issues}, "", "  ")
		fmt.Println(string(out))
	} else {
		printConfigIssues(os.Stderr, path, issues)
		if len(issues) == 0 {
			fmt.Printf("✓ %s matches the krewire.yaml schema\n", path)
		}
	}
	if config.HasErrors(issues) || (flagBool(fs, "strict") && len(issues) > 0) {
		return errs.ExitCodeFailure
	}
	return errs.ExitCodeSuccess
}

// printConfigIssues prints issues as path:line:col: severity: message.
func printConfigIssues(w io.Writer, path string, issues []config.Issue) {
	for _, i := range issues {
		fmt.Fprintf(w, "%s:%s\n", path, i)
	}
}

// validateProjectConfig validates dir/krewire.yaml when present. It reports
// whether the file exists and the issues found.
func validateProjectConfig(dir string) (bool, []config.Issue, error) {
	path := filepath.Join(dir, "krewire.yaml")
	if _, err := os.Stat(path); err != nil {
		return false, nil, nil
	}
	issues, err := config.ValidateFile(path)
	return true, issues, err
}
