package main

import (
	"os"

	"github.com/krewire/krew/internal/cli"
)

func main() {
	app := cli.BuildApp("krew")
	os.Exit(int(app.Run(os.Args[1:])))
}
