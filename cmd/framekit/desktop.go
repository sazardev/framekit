package main

import (
	"flag"

	"github.com/sazardev/framekit/desktop"
)

func desktopCmd(args []string) {
	fs := flag.NewFlagSet("desktop", flag.ExitOnError)
	port := fs.Int("port", 8239, "local port")
	noOpen := fs.Bool("no-open", false, "do not open the browser")
	_ = fs.Parse(args)
	if err := desktop.Run(*port, *noOpen); err != nil {
		fatal(err)
	}
}
