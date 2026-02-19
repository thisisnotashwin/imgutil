package main

import (
	"fmt"
	"os"

	"github.com/thisisnotashwin/imgutil/commands"
	"github.com/thisisnotashwin/imgutil/internal/image"
)

func main() {
	loader := image.NewLoader()
	if err := commands.NewRootCmd(loader).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
