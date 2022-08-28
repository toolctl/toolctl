// Package main only calls `cmd.Execute()`, which is the entry point for the CLI.
package main

import (
	"github.com/toolctl/toolctl/internal/cmd"
)

func main() {
	cmd.Execute()
}
