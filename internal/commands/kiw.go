package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"

	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/term"
)

// ClientCatcalls contains humorous client-to-developer catcalling quotes.
var ClientCatcalls = []string{
	"Hey handsome, is that a tight deadline in your pocket, or are you just excited to overhaul the architecture by tomorrow morning?",
	"Damn dev, you're looking like a single static binary with zero external dependencies today!",
	"Hey gorgeous, are you an urgent hotfix? Because you've been keeping me up all night.",
	"Psst cutie, mind if I creep into your scope with just one tiny, budgetless change request?",
	"Damn, look at those sub-millisecond cold starts! You got an API documentation for that body?",
	"Hey rockstar, does your SLA cover falling head over heels for a client like me?",
	"Hey beautiful, can you deploy straight to my production on a Friday night without a rollback plan?",
	"Looking sharp, coder! How about we skip the staging environment and merge our branches tonight?",
	"Damn babe, are you a microservice? Because you're looking distributed, resilient, and impossible to replace.",
	"Hey sweetie, I've got zero budget and unlimited requirements, and you're the only 10x engineer who can handle me.",
	"Hey sexy, are you CORS? Because you're allowing all origins into my heart.",
	"Damn dev, is your code compiled? Because every time you look at me, you resolve all my merge conflicts.",
	"Hey cutie, forget the backlog... let's prioritize each other in this sprint.",
	"Excuse me, engineer, are you a zero-downtime deployment? Because you just took my breath away without interrupting service.",
	"Hey gorgeous, you must be pure Go stdlib, because you give me zero JS fatigue and 100% satisfaction.",
	"Hey dev, is your cache invalidated? Because I simply can't get you out of my memory.",
}

var (
	kiwFlagAll  bool
	kiwFlagJSON bool
)

// RegisterKiw configures flags for the `kiw kiw` easter egg command.
func RegisterKiw(fs *flag.FlagSet) {
	fs.BoolVar(&kiwFlagAll, "all", false, "display all client-to-developer catcalling quotes")
	fs.BoolVar(&kiwFlagJSON, "json", false, "output catcalling quote(s) in JSON format")
}

// RunKiw executes the `kiw kiw` easter egg command.
func RunKiw(_ *flag.FlagSet) errs.ExitCode {
	if kiwFlagJSON {
		if kiwFlagAll {
			data, _ := json.MarshalIndent(map[string]any{
				"tag":      "kiw kiw",
				"total":    len(ClientCatcalls),
				"catcalls": ClientCatcalls,
			}, "", "  ")
			fmt.Println(string(data))
			return errs.ExitCodeSuccess
		}
		quote := ClientCatcalls[rand.IntN(len(ClientCatcalls))]
		data, _ := json.MarshalIndent(map[string]string{
			"tag":     "kiw kiw",
			"catcall": quote,
			"source":  "Client to Developer",
		}, "", "  ")
		fmt.Println(string(data))
		return errs.ExitCodeSuccess
	}

	tm := term.NewTerminal()
	cyanBold := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	magentaBold := func(s string) string { return tm.Paint(s, term.ColorMagenta, []term.Style{term.StyleBold}) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, nil) }
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }

	if kiwFlagAll {
		fmt.Printf("\n%s %s\n\n", magentaBold("(˵ •̀ ᴗ - ˵ ) ✧"), cyanBold("kiw kiw! — All Client Catcalls"))
		for i, q := range ClientCatcalls {
			fmt.Printf("  %s %s\n", dim(fmt.Sprintf("[%02d]", i+1)), yellow(fmt.Sprintf("%q", q)))
		}
		fmt.Printf("\n  %s\n\n", dim("— Playfully whispered from Clients to Developers"))
		return errs.ExitCodeSuccess
	}

	quote := ClientCatcalls[rand.IntN(len(ClientCatcalls))]
	fmt.Printf("\n%s %s\n\n", magentaBold("(˵ •̀ ᴗ - ˵ ) ✧"), cyanBold("kiw kiw!"))
	fmt.Printf("  %s\n\n", yellow(fmt.Sprintf("%q", quote)))
	fmt.Printf("  %s\n\n", dim("— Client to Developer"))

	return errs.ExitCodeSuccess
}
