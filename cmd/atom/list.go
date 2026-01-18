package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/scottwater/atoms/internal/storage"
	"github.com/scottwater/atoms/internal/task"
	"github.com/spf13/cobra"
)

var (
	listStatus   string
	listType     string
	listPriority int
	listParent   string
	listJSON     bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  `List all tasks with optional filtering by status, type, priority, or parent.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (open, in_progress, blocked, closed)")
	listCmd.Flags().StringVar(&listType, "type", "", "Filter by type (feature, bug)")
	listCmd.Flags().IntVar(&listPriority, "priority", 0, "Filter by priority (1, 2, or 3)")
	listCmd.Flags().StringVar(&listParent, "parent", "", "Show children of task")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Apply filters
	filtered := filterTasks(tasks)

	if listJSON {
		return outputJSON(filtered)
	}

	return outputTable(filtered)
}

func filterTasks(tasks []*task.Task) []*task.Task {
	var result []*task.Task

	for _, t := range tasks {
		// Filter by status
		if listStatus != "" && string(t.Status) != strings.ToLower(listStatus) {
			continue
		}

		// Filter by type
		if listType != "" && string(t.Type) != strings.ToLower(listType) {
			continue
		}

		// Filter by priority
		if listPriority != 0 && t.Priority != listPriority {
			continue
		}

		// Filter by parent
		if listParent != "" && t.ParentID != listParent {
			continue
		}

		// Default: show non-closed tasks if no status filter specified
		if listStatus == "" && t.Status == task.StatusClosed {
			continue
		}

		result = append(result, t)
	}

	return result
}

func outputJSON(tasks []*task.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputTable(tasks []*task.Task) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	// Print header
	fmt.Printf("%-12s %-4s %-8s %-12s %s\n", "ID", "PRI", "TYPE", "STATUS", "TITLE")

	for _, t := range tasks {
		fmt.Printf("%-12s P%-3d %-8s %-12s %s\n",
			t.ID,
			t.Priority,
			t.Type,
			t.Status,
			truncate(t.Title, 50),
		)
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
