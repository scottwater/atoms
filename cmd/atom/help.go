package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help and available commands",
	Long:  `Display detailed help information about atoms and available commands.`,
	Run:   runHelp,
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

func runHelp(cmd *cobra.Command, args []string) {
	fmt.Println(`atoms - Minimal git-backed task tracker

USAGE:
  atom <command> [flags]

SETUP COMMANDS:
  init        Initialize atoms in current directory
  help        Show this help message
  onboard     Display ATOM.md content for AI agents

TASK COMMANDS:
  create      Create a new task
  list        List all tasks (with optional filters)
  show        Show task details
  update      Update task fields
  close       Close a task
  ready       List tasks ready for work (open, not blocked)

GIT COMMANDS:
  merge       Git merge driver (called by git during merges)

EXAMPLES:
  # Initialize atoms in a project
  atom init

  # Create a new feature task
  atom create "Add user authentication" --type feature --priority 1

  # Create a bug with description
  atom create "Fix login validation" --type bug -d "Users can't log in with special chars"

  # List all open tasks
  atom list --status open

  # Find ready work
  atom ready

  # Update task status
  atom update atom-a3f2 --status in_progress

  # Close completed task
  atom close atom-a3f2

  # Get JSON output for scripts/AI
  atom list --json
  atom ready --json

WORKFLOW:
  1. Run 'atom ready' to find available work
  2. Claim work: 'atom update <id> --status in_progress'
  3. Do the work
  4. Complete: 'atom close <id>'
  5. Commit: 'git add .atoms.jsonl && git commit'

For more information on a command: atom <command> --help`)
}
