package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/scottwater/atoms/internal/storage"
	"github.com/spf13/cobra"
)

var initPrefix string
var initQuiet bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize atoms in current directory",
	Long:  `Creates .atoms.jsonl, .gitattributes, configures git merge driver, and creates ATOM.md.`,
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&initPrefix, "prefix", "atom", "Custom ID prefix")
	initCmd.Flags().BoolVar(&initQuiet, "quiet", false, "Suppress output")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create .atoms.jsonl
	store := storage.New(dir)
	if !store.Exists() {
		if err := store.Create(); err != nil {
			return fmt.Errorf("failed to create storage file: %w", err)
		}
	}

	// Create/update .gitattributes
	if err := createGitAttributes(dir); err != nil {
		return fmt.Errorf("failed to create .gitattributes: %w", err)
	}

	// Configure git merge driver
	if err := configureGitMergeDriver(); err != nil {
		return fmt.Errorf("failed to configure git merge driver: %w", err)
	}

	// Create ATOM.md if it doesn't exist
	if err := createAtomMD(dir); err != nil {
		return fmt.Errorf("failed to create ATOM.md: %w", err)
	}

	if !initQuiet {
		fmt.Println("✓ atoms initialized successfully!")
		fmt.Println()
		fmt.Println("  Storage: .atoms.jsonl")
		fmt.Printf("  Task prefix: %s\n", initPrefix)
		fmt.Println()
		fmt.Println("Run 'atom onboard' to see ATOM.md content for AI agents.")
	}

	return nil
}

func createGitAttributes(dir string) error {
	path := filepath.Join(dir, ".gitattributes")
	entry := ".atoms.jsonl merge=atoms\n"

	// Check if file exists and already has the entry
	if data, err := os.ReadFile(path); err == nil {
		content := string(data)
		if contains(content, ".atoms.jsonl merge=atoms") {
			return nil
		}
		// Append to existing file
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(entry)
		return err
	}

	// Create new file
	return os.WriteFile(path, []byte(entry), 0644)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func configureGitMergeDriver() error {
	// Check if we're in a git repository
	if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
		return nil // Not a git repo, skip configuration
	}

	// Configure merge driver
	if err := exec.Command("git", "config", "merge.atoms.driver", "atom merge %O %A %B").Run(); err != nil {
		return fmt.Errorf("failed to set merge driver: %w", err)
	}

	if err := exec.Command("git", "config", "merge.atoms.name", "atoms JSONL merge driver").Run(); err != nil {
		return fmt.Errorf("failed to set merge driver name: %w", err)
	}

	return nil
}

func createAtomMD(dir string) error {
	path := filepath.Join(dir, "ATOM.md")

	// Don't overwrite existing file
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	content := `# Task Tracking

This project uses **atom** for lightweight task tracking.

Run ` + "`atom ready`" + ` to see available work, or ` + "`atom help`" + ` for all commands.

## Quick Reference

` + "```bash" + `
atom ready              # Find available work
atom show <id>          # View task details  
atom update <id> --status in_progress  # Claim work
atom close <id>         # Complete work
` + "```" + `

## Session Completion

When ending work:
1. Close completed tasks: ` + "`atom close <id>`" + `
2. Commit changes: ` + "`git add .atoms.jsonl && git commit`" + `
3. Push to remote: ` + "`git push`" + `
`

	return os.WriteFile(path, []byte(content), 0644)
}
