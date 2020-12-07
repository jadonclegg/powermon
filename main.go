package main

import (
	"powermon/commands"

	"github.com/jessevdk/go-flags"
)

func main() {
	flags.Parse(&commands.Powermon)
}
