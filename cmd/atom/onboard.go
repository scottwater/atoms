package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Display ATOM.md content for AI agents",
	Long:  `Output content suitable for ATOM.md to help AI agents understand task tracking.`,
	Run:   runOnboard,
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}

func runOnboard(cmd *cobra.Command, args []string) {
	fmt.Println(`Add this to AGENTS.md or ATOM.md:

# Task Tracking

This project uses **atom** for lightweight task tracking.

Run ` + "`atom ready`" + ` to see available work, or ` + "`atom help`" + ` for all commands.

## Quick Reference

` + "```bash" + `
atom ready              # Find available work
atom show <id>          # View task details  
atom update <id> --status in_progress  # Claim work
atom close <id>         # Complete work
` + "```" + `

## Workflow

1. Check for ready work: ` + "`atom ready`" + `
2. Claim your task: ` + "`atom update <id> --status in_progress`" + `
3. Do the work
4. Complete: ` + "`atom close <id>`" + `
5. Commit: ` + "`git add .atoms.jsonl && git commit`" + ``)
}
