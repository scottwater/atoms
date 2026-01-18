package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/scottwater/atoms/internal/storage"
	"github.com/scottwater/atoms/internal/task"
	"github.com/spf13/cobra"
)

var readyJSON bool

var readyCmd = &cobra.Command{
	Use:   "ready",
	Short: "List tasks ready for work",
	Long:  `List tasks that are ready for work (open status, not blocked).`,
	RunE:  runReady,
}

func init() {
	readyCmd.Flags().BoolVar(&readyJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(readyCmd)
}

func runReady(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	store := storage.New(dir)
	if !store.Exists() {
		return fmt.Errorf("atoms not initialized. Run 'atom init' first")
	}

	tasks, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	// Filter to open tasks only (not in_progress, blocked, or closed)
	var ready []*task.Task
	for _, t := range tasks {
		if t.Status == task.StatusOpen {
			ready = append(ready, t)
		}
	}

	if readyJSON {
		return outputReadyJSON(ready)
	}

	return outputReadyTable(ready)
}

func outputReadyJSON(tasks []*task.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputReadyTable(tasks []*task.Task) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks ready for work.")
		return nil
	}

	// Print header
	fmt.Printf("%-12s %-4s %-8s %s\n", "ID", "PRI", "TYPE", "TITLE")

	for _, t := range tasks {
		fmt.Printf("%-12s P%-3d %-8s %s\n",
			t.ID,
			t.Priority,
			t.Type,
			truncate(t.Title, 50),
		)
	}

	return nil
}
