// Command framekit-tui launches the interactive terminal UI.
package main

import (
	"fmt"
	"os"

	"github.com/sazardev/framekit/tui"
)

func main() {
	if err := tui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
