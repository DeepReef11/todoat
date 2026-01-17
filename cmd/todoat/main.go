package main

import (
	"os"

	cmd "todoat/cmd/todoat/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:], os.Stdout, os.Stderr, nil))
}
