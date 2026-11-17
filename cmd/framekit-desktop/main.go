// Command framekit-desktop opens the local web studio.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sazardev/framekit/desktop"
)

func main() {
	port := flag.Int("port", 8239, "local port")
	noOpen := flag.Bool("no-open", false, "do not open the browser")
	flag.Parse()
	if err := desktop.Run(*port, *noOpen); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
