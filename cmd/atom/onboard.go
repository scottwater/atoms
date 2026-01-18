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
	fmt.Println(`
Add this to ATOM.md (or create it):

--- BEGIN ATOM.MD CONTENT ---
## Task Tracking

This project uses **atom** for task tracking.

**Quick reference:**
- ` + "`atom ready`" + ` - Find available work
- ` + "`atom create \"Title\" --type feature --priority 2`" + ` - Create task
- ` + "`atom close <id>`" + ` - Complete work
- ` + "`atom list`" + ` - See all tasks

**Workflow:**
1. Run ` + "`atom ready`" + ` to find work
2. Update status: ` + "`atom update <id> --status in_progress`" + `
3. Do the work
4. Close: ` + "`atom close <id>`" + `
5. Commit ` + "`.atoms.jsonl`" + ` with your changes
--- END ATOM.MD CONTENT ---`)
}
