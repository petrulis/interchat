package main

import (
	"github.com/petrulis/interchat/interchat"
	"github.com/petrulis/interchat/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Start chat service",
		Long:  `Runs the chat service`,
		RunE:  run,
	}
)

func run(cmd *cobra.Command, args []string) error {
	logrus.Info(version.Get())
	return interchat.FromEnv().Run()
}

func init() {
	cmd.AddCommand(runCmd)
}
