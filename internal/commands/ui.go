package commands

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/krewire/krew/internal/kui"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/ui/components"
)

// RegisterUI registers flags for the ui command.
func RegisterUI(fs *flag.FlagSet) {
	fs.Bool("force", false, "add: overwrite locally modified components")
	fs.Bool("dry-run", false, "add: show what would change without writing")
	fs.Bool("json", false, "list/info: print JSON")
}

const uiUsage = `usage:
  krew ui list                      catalog, starter packs, and installed state
  krew ui info <name>               props, defaults, dependencies, preview
  krew ui vendor <name|ui:pack> ... copy into ./components (explicit local ownership)
  krew ui eject <name|ui:pack> ...  alias for vendor
  krew ui add <name|ui:pack> ...    copy into ./components (same as krew add ui/<name>)
  krew ui diff <name>               diff upstream kui -> your local copy`

// RunUI dispatches the kui component subcommands.
func RunUI(fs *flag.FlagSet) errs.ExitCode {
	args := fs.Args()
	if len(args) == 0 {
		args = []string{"list"}
	}
	root, err := findRoot()
	if err != nil {
		root, _ = os.Getwd()
	}
	switch args[0] {
	case "list", "ls":
		return uiList(root, flagBool(fs, "json"))
	case "info", "show":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, uiUsage)
			return errs.ExitCodeUsage
		}
		return uiInfo(root, args[1], flagBool(fs, "json"))
	case "add", "install", "vendor", "eject":
		return uiAdd(root, args[1:], flagBool(fs, "force"), flagBool(fs, "dry-run"))
	case "diff":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, uiUsage)
			return errs.ExitCodeUsage
		}
		return uiDiff(root, args[1])
	default:
		fmt.Fprintf(os.Stderr, "krew ui: unknown subcommand %q\n%s\n", args[0], uiUsage)
		return errs.ExitCodeUsage
	}
}

// isUIRef reports whether a krew add argument names kui components.
func isUIRef(arg string) bool {
	return strings.HasPrefix(arg, "ui/") || strings.HasPrefix(arg, "ui:")
}

func uiAdd(root string, refs []string, force, dryRun bool) errs.ExitCode {
	if len(refs) == 0 {
		fmt.Fprintln(os.Stderr, uiUsage)
		fmt.Fprintln(os.Stderr, "hint: run 'krew ui list' to browse components and packs")
		return errs.ExitCodeUsage
	}
	results, err := kui.Install(root, refs, force, dryRun)
	if err != nil && !errors.Is(err, kui.ErrModified) {
		fmt.Fprintln(os.Stderr, "krew ui: "+err.Error())
		fmt.Fprintln(os.Stderr, "hint: run 'krew ui list' to see component and pack names")
		return errs.ExitCodeUsage
	}
	prefix := ""
	if dryRun {
		prefix = "(dry run) "
	}
	for _, r := range results {
		mark := "✓"
		if r.Action == kui.ActionKept {
			mark = "!"
		}
		fmt.Printf("%s %s%s %s\n", mark, prefix, filepath.ToSlash(filepath.Join(kui.Dir, r.Name+".kiw")), r.Action)
	}
	if err != nil {
		for _, r := range results {
			if r.Action == kui.ActionKept {
				fmt.Fprintf(os.Stderr, "krew ui: %s has local changes — review with 'krew ui diff %s', replace with --force\n", r.Name, r.Name)
			}
		}
		return errs.ExitCodeFailure
	}
	return errs.ExitCodeSuccess
}

func uiList(root string, asJSON bool) errs.ExitCode {
	lock, err := kui.ReadLock(root)
	if err != nil {
		return fail(err)
	}
	type row struct {
		components.ComponentMeta
		Status kui.Status `json:"status"`
	}
	byCat := map[string][]row{}
	var all []row
	for _, name := range components.List() {
		st, err := kui.State(root, name, lock)
		if err != nil {
			return fail(err)
		}
		r := row{components.Registry[name], st}
		byCat[r.Category] = append(byCat[r.Category], r)
		all = append(all, r)
	}
	if asJSON {
		out, _ := json.MarshalIndent(map[string]any{"kui": components.Version(), "packs": components.Packs(), "components": all}, "", "  ")
		fmt.Println(string(out))
		return errs.ExitCodeSuccess
	}
	cats := make([]string, 0, len(byCat))
	for c := range byCat {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	fmt.Printf("kui %s — %d components\n", components.Version(), len(all))
	for _, c := range cats {
		fmt.Printf("\n%s\n", strings.ToUpper(c))
		for _, r := range byCat[c] {
			state := ""
			if r.Status != kui.StatusMissing {
				state = "  [" + string(r.Status) + "]"
			}
			fmt.Printf("  %-15s %s%s\n", r.Name, r.Description, state)
		}
	}
	fmt.Println("\nSTARTER PACKS")
	for _, p := range components.Packs() {
		fmt.Printf("  ui:%-12s %s\n", p.Name, strings.Join(p.Components, ", "))
	}
	fmt.Println("\nAdd with: krew add ui/<name> or krew add ui:<pack>   Details: krew ui info <name>")
	return errs.ExitCodeSuccess
}

func uiInfo(root, name string, asJSON bool) errs.ExitCode {
	c, ok := components.Lookup(strings.TrimPrefix(name, "ui/"))
	if !ok {
		fmt.Fprintf(os.Stderr, "krew ui: unknown component %q — run 'krew ui list'\n", name)
		return errs.ExitCodeUsage
	}
	lock, err := kui.ReadLock(root)
	if err != nil {
		return fail(err)
	}
	st, err := kui.State(root, c.Name, lock)
	if err != nil {
		return fail(err)
	}
	if asJSON {
		out, _ := json.MarshalIndent(map[string]any{"component": c, "status": st}, "", "  ")
		fmt.Println(string(out))
		return errs.ExitCodeSuccess
	}
	fmt.Printf("%s (%s) — %s\n", c.Name, c.Category, c.Description)
	fmt.Printf("  status:       %s\n", st)
	fmt.Printf("  props:        %s\n", strings.Join(c.Props, ", "))
	if len(c.Defaults) > 0 {
		var d []string
		for k, v := range c.Defaults {
			d = append(d, k+"="+v)
		}
		sort.Strings(d)
		fmt.Printf("  defaults:     %s\n", strings.Join(d, ", "))
	}
	if len(c.Dependencies) > 0 {
		fmt.Printf("  dependencies: %s (installed together)\n", strings.Join(c.Dependencies, ", "))
	}
	fmt.Printf("  preview:      %s\n", c.Preview)
	fmt.Printf("\nAdd with: krew add ui/%s\n", strings.ToLower(c.Name))
	return errs.ExitCodeSuccess
}

func uiDiff(root, name string) errs.ExitCode {
	c, ok := components.Lookup(strings.TrimPrefix(name, "ui/"))
	if !ok {
		fmt.Fprintf(os.Stderr, "krew ui: unknown component %q — run 'krew ui list'\n", name)
		return errs.ExitCodeUsage
	}
	local := kui.Path(root, c.Name)
	if _, err := os.Stat(local); err != nil {
		fmt.Fprintf(os.Stderr, "krew ui: %s is not installed — add it with 'krew add ui/%s'\n", filepath.ToSlash(filepath.Join(kui.Dir, c.Name+".kiw")), strings.ToLower(c.Name))
		return errs.ExitCodeUsage
	}
	upstream, err := components.Get(c.Name)
	if err != nil {
		return fail(err)
	}
	tmp, err := os.MkdirTemp("", "kiw-ui-diff-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(tmp)
	up := filepath.Join(tmp, "kui-"+components.Version(), c.Name+".kiw")
	if err := os.MkdirAll(filepath.Dir(up), 0o755); err != nil {
		return fail(err)
	}
	if err := os.WriteFile(up, upstream, 0o644); err != nil {
		return fail(err)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		fmt.Fprintf(os.Stderr, "krew ui: git not found; compare %s with %s manually\n", up, local)
		return errs.ExitCodeFailure
	}
	cmd := exec.Command(git, "diff", "--no-index", "--", up, local) // #nosec G204 — fixed argv
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return errs.ExitCodeSuccess // differences found
		}
		return fail(err)
	}
	fmt.Printf("%s matches kui %s\n", c.Name, components.Version())
	return errs.ExitCodeSuccess
}
