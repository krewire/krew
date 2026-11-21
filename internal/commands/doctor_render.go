package commands

import (
	"fmt"
	"strings"

	"github.com/krewire/krewire/pkg/term"
)

func renderDoctorConsole(report DoctorReport) {
	tm := term.NewTerminal()
	bold := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, []term.Style{term.StyleBold}) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, []term.Style{term.StyleBold}) }
	red := func(s string) string { return tm.Paint(s, term.ColorRed, []term.Style{term.StyleBold}) }
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }

	fmt.Println()
	fmt.Printf("%s\n", bold("Krewire Environment Doctor"))
	fmt.Println(dim(strings.Repeat("─", 50)))
	fmt.Printf("OS/Arch:    %s/%s (%s CPUs)\n", report.Environment["os"], report.Environment["arch"], report.Environment["cpus"])
	fmt.Printf("Krew CLI:   %s\n", report.Environment["krew_version"])
	if cr, ok := report.Environment["go_compiler"]; ok {
		fmt.Printf("Go runtime: %s\n", cr)
	}
	fmt.Println(dim(strings.Repeat("─", 50)))
	fmt.Println()

	for _, c := range report.Checks {
		var badge string
		switch c.Status {
		case StatusPass:
			badge = green("[ PASS ]")
		case StatusWarn:
			badge = yellow("[ WARN ]")
		case StatusFail:
			badge = red("[ FAIL ]")
		}

		fmt.Printf("%s  %s: %s\n", badge, bold(c.Name), c.Message)
		if c.Remediation != "" {
			fmt.Printf("          %s %s\n", dim("↳ Fix:"), c.Remediation)
		}
	}

	fmt.Println()
	fmt.Println(dim(strings.Repeat("─", 50)))
	if report.Passed {
		fmt.Printf("%s %s\n", green("Doctor Result:"), report.Summary)
	} else {
		fmt.Printf("%s %s\n", red("Doctor Result:"), report.Summary)
	}
	fmt.Println()
}
