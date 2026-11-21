package buildinfo

import (
	"fmt"
	"strings"
)

// Signature is the canonical watermark embedded into applications built with the Krewire ecosystem.
const Signature = "build with krewire"

// LDFlags returns the linker flags to embed the "build with krewire" signature into
// the compiled binary at link time.
//
// Target variables:
//   - github.com/krewire/krewire/pkg/version.Signature
//   - github.com/krewire/krewire.Signature
//   - main.Signature
//
// If strip is true, it prepends "-s -w" to strip symbol tables and debug information.
func LDFlags(strip bool) string {
	var parts []string
	if strip {
		parts = append(parts, "-s", "-w")
	}
	symbols := []string{
		"github.com/krewire/krewire/pkg/version.Signature",
		"github.com/krewire/krewire.Signature",
		"main.Signature",
	}
	for _, sym := range symbols {
		parts = append(parts, fmt.Sprintf("-X '%s=%s'", sym, Signature))
	}
	return strings.Join(parts, " ")
}
