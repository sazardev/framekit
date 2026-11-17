package main

import (
	"flag"

	"github.com/sazardev/framekit/tui"
)

func tuiCmd(args []string) {
	_ = flag.NewFlagSet("tui", flag.ExitOnError).Parse(args)
	if err := tui.Run(); err != nil {
		fatal(err)
	}
}
