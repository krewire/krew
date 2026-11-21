package commands

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/krewire/krewire/pkg/term"
)

// projectKind is one starter offered by `krew new` and `krew init`; Flag is the
// boolean flag both commands register for it.
type projectKind struct {
	Flag, Label string
}

// projectKinds lists the starters in the order they are offered.
var projectKinds = []projectKind{
	{"app", "fullstack web app (Go server + pages)"},
	{"site", "static site (pages/, layouts/, components/)"},
	{"book", "documentation book from Markdown in content/"},
	{"cli", "command-line tool"},
	{"worker", "background job worker"},
	{"service", "standalone microservice"},
	{"gateway", "API gateway"},
	{"infra", "cloud infrastructure blueprint"},
}

// Prompt I/O and terminal detection, replaceable in tests.
var (
	promptIn    io.Reader = os.Stdin
	promptOut   io.Writer = os.Stdout
	interactive           = func() bool {
		if !term.IsTerminal(os.Stdin) || !term.IsTerminal(os.Stdout) {
			return false
		}
		// /dev/null is a character device too; it never answers.
		in, err := os.Stdin.Stat()
		if err != nil {
			return false
		}
		null, err := os.Stat(os.DevNull)
		return err != nil || !os.SameFile(in, null)
	}
)

// shouldPrompt reports whether a command may ask questions: stdin is a
// terminal and --no-input was not given.
func shouldPrompt(fs *flag.FlagSet) bool {
	return !flagBool(fs, "no-input") && interactive()
}

// ask prints question and returns the trimmed answer, or def when empty.
func ask(r *bufio.Reader, question, def string) string {
	if def != "" {
		fmt.Fprintf(promptOut, "%s [%s]: ", question, def)
	} else {
		fmt.Fprintf(promptOut, "%s: ", question)
	}
	line, _ := r.ReadString('\n')
	if line = strings.TrimSpace(line); line != "" {
		return line
	}
	return def
}

// askKind offers the starters and returns the chosen flag name, or "" for
// none when allowNone is set. It accepts a number or a name.
func askKind(r *bufio.Reader, def string, allowNone bool) string {
	fmt.Fprintln(promptOut, "What do you want to build?")
	for i, k := range projectKinds {
		fmt.Fprintf(promptOut, "  %d) %-8s %s\n", i+1, k.Flag, k.Label)
	}
	if allowNone {
		fmt.Fprintf(promptOut, "  0) %-8s only krewire.yaml, .gitignore and git (equip later)\n", "none")
	}
	for attempt := 0; attempt < 3; attempt++ {
		ans := strings.ToLower(ask(r, "Kind", def))
		if allowNone && (ans == "0" || ans == "none") {
			return ""
		}
		if n, err := strconv.Atoi(ans); err == nil && n >= 1 && n <= len(projectKinds) {
			return projectKinds[n-1].Flag
		}
		for _, k := range projectKinds {
			if ans == k.Flag {
				return k.Flag
			}
		}
		fmt.Fprintf(promptOut, "  %q is not a kind; enter a number or a name\n", ans)
	}
	return def
}

// hasKindFlag reports whether any starter flag or --template is set.
func hasKindFlag(fs *flag.FlagSet) bool {
	for _, k := range projectKinds {
		if flagBool(fs, k.Flag) {
			return true
		}
	}
	return flagValue(fs, "template") != ""
}
