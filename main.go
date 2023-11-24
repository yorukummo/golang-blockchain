// Package main is used to run the program.
package main

import (
	"os"

	"github.com/argonautts/golang-blockchain/cli"
)

func main() {
	// We use defer to ensure that the program exits with code 0 when main() completes.
	defer os.Exit(0)

	cmd := cli.CommandLine{}
	cmd.Run()
}
