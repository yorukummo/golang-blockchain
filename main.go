package main

import (
	"os"

	"github.com/argonautts/golang-blockchain/cli"
)

// github.com/argonautts/golang-blockchain/blockchain/cli

func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()
}
