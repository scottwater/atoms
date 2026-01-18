package main

import (
	"fmt"
	"os"

	"github.com/scottwater/atoms/internal/storage"
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close [id]",
	Short: "Close a task",
	Long:  `Close a task by changing its status to closed.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runClose,
}

func init() {
	rootCmd.AddCommand(closeCmd)
}

func runClose(cmd *cobra.Command, args []string) error {
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

	t.Close()

	if err := store.Update(t); err != nil {
		return fmt.Errorf("failed to close task: %w", err)
	}

	fmt.Printf("✓ Closed %s: %s\n", t.ID, t.Title)

	return nil
}
