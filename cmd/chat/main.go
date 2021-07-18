package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:          fmt.Sprintf("%s [command]", os.Args[0]),
		Short:        "Chat service",
		Long:         `The chat service enables bi-directional communication over websockets`,
		SilenceUsage: true,
	}
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
