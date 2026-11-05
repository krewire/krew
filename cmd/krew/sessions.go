package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/krewire/krew/pkg/memory"
)

func printSessionList(w io.Writer, sessions []memory.SessionInfo) {
	fmt.Fprintln(w, "Saved Sessions:")
	for _, s := range sessions {
		fmt.Fprintf(w, "  • %-36s  %s  (%d msgs, %d tokens)\n",
			s.ID,
			s.ModTime.Format("2006-01-02 15:04:05"),
			s.Messages,
			s.Tokens,
		)
	}
}

func runSessions(args []string) {
	fs := flag.NewFlagSet("sessions", flag.ExitOnError)
	dirFlag := fs.String("dir", "", "Custom sessions directory")
	_ = fs.Parse(args)

	sessDir := *dirFlag
	if sessDir == "" {
		var err error
		sessDir, err = memory.DefaultSessionsDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error locating sessions directory: %v\n", err)
			os.Exit(1)
		}
	}

	sessions, err := memory.ListSessions(sessDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No recorded sessions found.")
		return
	}

	printSessionList(os.Stdout, sessions)
}

func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	formatFlag := fs.String("format", DefaultExportFormat, "Export format (default: markdown)")
	dirFlag := fs.String("dir", "", "Custom sessions directory")
	_ = fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "error: session ID required for 'krew export'")
		os.Exit(2)
	}
	sessionID := remaining[0]

	sessDir := *dirFlag
	if sessDir == "" {
		var err error
		sessDir, err = memory.DefaultSessionsDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error locating sessions directory: %v\n", err)
			os.Exit(1)
		}
	}

	if *formatFlag != "markdown" && *formatFlag != "md" {
		fmt.Fprintf(os.Stderr, "unsupported export format %q (supported: markdown)\n", *formatFlag)
		os.Exit(2)
	}

	rendered, err := memory.RenderTranscriptMarkdown(sessDir, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting session %s: %v\n", sessionID, err)
		os.Exit(1)
	}

	fmt.Print(rendered)
}
