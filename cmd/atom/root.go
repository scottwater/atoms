package main

import (
	"github.com/spf13/cobra"
)

var Version = "0.5.0"

var rootCmd = &cobra.Command{
	Use:     "atom",
	Short:   "Minimal git-backed task tracker",
	Long:    `Atoms is a lightweight, portable task tracker designed for AI agents and developers.`,
	Version: Version,
}

func Execute() error {
	return rootCmd.Execute()
}
