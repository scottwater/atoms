package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/scottwater/atoms/internal/storage"
	"github.com/scottwater/atoms/internal/task"
	"github.com/spf13/cobra"
)

var showJSON bool

var showCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show task details",
	Long:  `Display detailed information about a task, including its children.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
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

	// Find children
	allTasks, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	var children []*task.Task
	for _, child := range allTasks {
		if child.ParentID == id {
			children = append(children, child)
		}
	}

	if showJSON {
		return outputShowJSON(t, children)
	}

	return outputShowText(t, children)
}

func outputShowJSON(t *task.Task, children []*task.Task) error {
	output := struct {
		*task.Task
		Children []*task.Task `json:"children,omitempty"`
	}{
		Task:     t,
		Children: children,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputShowText(t *task.Task, children []*task.Task) error {
	fmt.Printf("ID:          %s\n", t.ID)
	fmt.Printf("Title:       %s\n", t.Title)
	fmt.Printf("Type:        %s\n", t.Type)
	fmt.Printf("Priority:    P%d\n", t.Priority)
	fmt.Printf("Status:      %s\n", t.Status)
	fmt.Printf("Created:     %s\n", t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	fmt.Printf("Updated:     %s\n", t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if t.CreatedBy != "" {
		fmt.Printf("Created By:  %s\n", t.CreatedBy)
	}

	if t.ParentID != "" {
		fmt.Printf("Parent:      %s\n", t.ParentID)
	}

	if t.Description != "" {
		fmt.Printf("Description: %s\n", t.Description)
	}

	if len(children) > 0 {
		fmt.Println()
		fmt.Println("Children:")
		for _, child := range children {
			fmt.Printf("  %-12s P%d  %-8s %-12s %s\n",
				child.ID,
				child.Priority,
				child.Type,
				child.Status,
				truncate(child.Title, 40),
			)
		}
	}

	return nil
}
