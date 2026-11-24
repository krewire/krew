package commands

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/krewire/tui/term"
)

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
