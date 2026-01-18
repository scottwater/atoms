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
	updateStatus      string
	updatePriority    int
	updateDescription string
	updateTitle       string
)

var updateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update task fields",
	Long:  `Update one or more fields of an existing task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateStatus, "status", "", "New status (open, in_progress, blocked, closed)")
	updateCmd.Flags().IntVarP(&updatePriority, "priority", "p", 0, "New priority (1, 2, or 3)")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New description")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	store := storage.New(dir)
	if !store.Exists() {
		return fmt.Errorf("atoms not initialized. Run 'atom init' first")
	}

	t, err := store.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}
	if t == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	updated := false

	// Update status
	if updateStatus != "" {
		status, err := parseStatus(updateStatus)
		if err != nil {
			return err
		}
		t.Status = status
		updated = true
	}

	// Update priority
	if updatePriority != 0 {
		if updatePriority < 1 || updatePriority > 3 {
			return fmt.Errorf("invalid priority: %d (must be 1, 2, or 3)", updatePriority)
		}
		t.Priority = updatePriority
		updated = true
	}

	// Update description
	if cmd.Flags().Changed("description") {
		t.Description = updateDescription
		updated = true
	}

	// Update title
	if updateTitle != "" {
		t.Title = updateTitle
		updated = true
	}

	if !updated {
		return fmt.Errorf("no updates specified")
	}

	t.Update()

	if err := store.Update(t); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Output JSON for agents
	output := map[string]string{
		"id":     t.ID,
		"title":  t.Title,
		"status": string(t.Status),
	}
	data, _ := json.Marshal(output)
	fmt.Println(string(data))

	return nil
}

func parseStatus(s string) (task.Status, error) {
	switch strings.ToLower(s) {
	case "open":
		return task.StatusOpen, nil
	case "in_progress":
		return task.StatusInProgress, nil
	case "blocked":
		return task.StatusBlocked, nil
	case "closed":
		return task.StatusClosed, nil
	default:
		return "", fmt.Errorf("invalid status: %s (must be open, in_progress, blocked, or closed)", s)
	}
}
