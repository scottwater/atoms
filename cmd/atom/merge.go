package main

import (
	"fmt"
	"os"

	"github.com/scottwater/atoms/internal/merge"
	"github.com/scottwater/atoms/internal/storage"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge [ancestor] [ours] [theirs]",
	Short: "Git merge driver for .atoms.jsonl",
	Long: `3-way merge driver for git. Called automatically by git during merge operations.

Configure with: git config merge.atoms.driver "atom merge %O %A %B"

Arguments:
  ancestor - path to common ancestor version (%O)
  ours     - path to our version (%A) - this file is overwritten with result
  theirs   - path to their version (%B)`,
	Args:   cobra.ExactArgs(3),
	RunE:   runMerge,
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) error {
	ancestorPath := args[0]
	oursPath := args[1]
	theirsPath := args[2]

	// Read all three versions
	ancestorStore := storage.NewWithPath(ancestorPath)
	ancestor, err := ancestorStore.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read ancestor: %w", err)
	}

	oursStore := storage.NewWithPath(oursPath)
	ours, err := oursStore.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read ours: %w", err)
	}

	theirsStore := storage.NewWithPath(theirsPath)
	theirs, err := theirsStore.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read theirs: %w", err)
	}

	// Perform 3-way merge
	merged := merge.Merge3Way(ancestor, ours, theirs)

	// Write result to ours path (git convention)
	if err := oursStore.WriteAll(merged); err != nil {
		return fmt.Errorf("failed to write merged result: %w", err)
	}

	// Exit 0 = merge successful
	os.Exit(0)
	return nil
}
