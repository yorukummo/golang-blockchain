package main

import (
	"os"

	"github.com/argonautts/golang-blockchain/cli"
)

func main() {
	// Используем defer, чтобы гарантировать выход из программы с кодом 0 при завершении main().
	defer os.Exit(0)

	cmd := cli.CommandLine{}
	cmd.Run()
}
