package main

import (
	"os"

	"github.com/tom-miy/ai-sensitive-files/internal/interface/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
