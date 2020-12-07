package main

import (
	"powermon/commands"

	"github.com/jessevdk/go-flags"
)

func main() {
	// flags.Parse will take in the os.Args, and call .Execute for the sepecified command.
	flags.Parse(&commands.Powermon)
}
