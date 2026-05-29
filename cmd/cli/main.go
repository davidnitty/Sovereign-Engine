package main

import (
	"fmt"
	"os"

	"github.com/yourusername/the-engine/internal/cli"
)

func main() {
	root := cli.NewRootCommand(cli.Options{})
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
