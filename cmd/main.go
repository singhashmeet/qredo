package main

import (
	"os"

	"github.com/singhashmeet/temp/cmd/qredo"

	"github.com/spf13/cobra"
)

// cmd is entrypoint for the program
var cmd = &cobra.Command{
	Use:   "qredo",
	Short: "API service",
}

func init() {
	// add DaemonCommand to cli
	cmd.AddCommand(qredo.DaemonCommand)
}

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
