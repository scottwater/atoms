package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/scottwater/atoms/internal/idgen"
	"github.com/scottwater/atoms/internal/storage"
	"github.com/scottwater/atoms/internal/task"
	"github.com/spf13/cobra"
)

var (
	createType        string
	createPriority    int
	createDescription string
	createParent      string
)

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new task",
	Long:  `Create a new task with the given title.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&createType, "type", "t", "feature", "Task type: feature or bug")
	createCmd.Flags().IntVarP(&createPriority, "priority", "p", 2, "Priority: 1 (highest) to 3 (lowest)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Detailed description")
	createCmd.Flags().StringVar(&createParent, "parent", "", "Parent task ID")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	title := args[0]

	// Validate type
	var taskType task.Type
	switch strings.ToLower(createType) {
	case "feature":
		taskType = task.TypeFeature
	case "bug":
		taskType = task.TypeBug
	default:
		return fmt.Errorf("invalid type: %s (must be 'feature' or 'bug')", createType)
	}

	// Validate priority
	if createPriority < 1 || createPriority > 3 {
		return fmt.Errorf("invalid priority: %d (must be 1, 2, or 3)", createPriority)
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	store := storage.New(dir)
	if !store.Exists() {
		return fmt.Errorf("atoms not initialized. Run 'atom init' first")
	}

	// Validate parent if specified
	if createParent != "" {
		parent, err := store.FindByID(createParent)
		if err != nil {
			return fmt.Errorf("failed to find parent task: %w", err)
		}
		if parent == nil {
			return fmt.Errorf("parent task not found: %s", createParent)
		}
	}

	// Create task
	t := task.New(title, taskType, createPriority)
	t.Description = createDescription
	t.ParentID = createParent

	// Get git user for created_by
	if out, err := exec.Command("git", "config", "user.name").Output(); err == nil {
		t.CreatedBy = strings.TrimSpace(string(out))
	}

	// Generate ID
	tasks, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read existing tasks: %w", err)
	}

	gen := idgen.New("atom")
	existingIDs := make([]string, len(tasks))
	for i, existing := range tasks {
		existingIDs[i] = existing.ID
	}
	gen.SetExisting(existingIDs)

	id, err := gen.Generate(title, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to generate ID: %w", err)
	}
	t.ID = id

	// Save task
	if err := store.Append(t); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
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
