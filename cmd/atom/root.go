package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "atom",
	Short: "Minimal git-backed task tracker",
	Long:  `Atoms is a lightweight, portable task tracker designed for AI agents and developers.`,
}

func Execute() error {
	return rootCmd.Execute()
}
