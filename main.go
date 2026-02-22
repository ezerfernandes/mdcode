package main

import (
	"os"

	"github.com/ezerfernandes/mdcode/internal/cmd"
)

//go:generate go run ./tools/gendoc ./README.md

func main() {
	cmd.Execute(os.Args[1:], os.Stdout, os.Stderr)
}
