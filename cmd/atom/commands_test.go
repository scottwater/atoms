package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottwater/atoms/internal/storage"
	"github.com/scottwater/atoms/internal/task"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "atom-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize as git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	cmd.Run()

	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestInitCreatesFiles(t *testing.T) {
	dir := setupTestDir(t)

	// Change to test directory
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	// Reset flags
	initPrefix = "atom"
	initQuiet = true

	// Run init
	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Check .atoms.jsonl exists
	if _, err := os.Stat(filepath.Join(dir, ".atoms.jsonl")); os.IsNotExist(err) {
		t.Error(".atoms.jsonl was not created")
	}

	// Check .gitattributes exists and has correct content
	data, err := os.ReadFile(filepath.Join(dir, ".gitattributes"))
	if err != nil {
		t.Fatalf("failed to read .gitattributes: %v", err)
	}
	if !strings.Contains(string(data), ".atoms.jsonl merge=atoms") {
		t.Error(".gitattributes missing merge driver config")
	}

	// Check ATOM.md exists
	if _, err := os.Stat(filepath.Join(dir, "ATOM.md")); os.IsNotExist(err) {
		t.Error("ATOM.md was not created")
	}
}

func TestInitDoesNotOverwriteAtomMD(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	// Create existing ATOM.md
	customContent := "# Custom Content"
	os.WriteFile(filepath.Join(dir, "ATOM.md"), []byte(customContent), 0644)

	initPrefix = "atom"
	initQuiet = true

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Check ATOM.md was not overwritten
	data, err := os.ReadFile(filepath.Join(dir, "ATOM.md"))
	if err != nil {
		t.Fatalf("failed to read ATOM.md: %v", err)
	}
	if string(data) != customContent {
		t.Error("ATOM.md was overwritten")
	}
}

func TestCreateTask(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	// Initialize first
	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	// Reset create flags
	createType = "bug"
	createPriority = 1
	createDescription = "Test description"
	createParent = ""

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(createCmd, []string{"Fix critical bug"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	if result["title"] != "Fix critical bug" {
		t.Errorf("expected title 'Fix critical bug', got '%s'", result["title"])
	}
	if result["status"] != "open" {
		t.Errorf("expected status 'open', got '%s'", result["status"])
	}
	if !strings.HasPrefix(result["id"], "atom-") {
		t.Errorf("expected id to start with 'atom-', got '%s'", result["id"])
	}

	// Verify task was saved
	store := storage.New(dir)
	tasks, _ := store.ReadAll()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Type != task.TypeBug {
		t.Errorf("expected type 'bug', got '%s'", tasks[0].Type)
	}
	if tasks[0].Priority != 1 {
		t.Errorf("expected priority 1, got %d", tasks[0].Priority)
	}
	if tasks[0].Description != "Test description" {
		t.Errorf("expected description 'Test description', got '%s'", tasks[0].Description)
	}
}

func TestCreateWithParent(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	// Create parent task
	createType = "feature"
	createPriority = 2
	createDescription = ""
	createParent = ""

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	runCreate(createCmd, []string{"Parent task"})
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	var result map[string]string
	json.Unmarshal([]byte(buf.String()), &result)
	parentID := result["id"]

	// Create child task
	createParent = parentID
	r, w, _ = os.Pipe()
	os.Stdout = w
	err := runCreate(createCmd, []string{"Child task"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("create child failed: %v", err)
	}

	// Verify parent_id was set
	store := storage.New(dir)
	tasks, _ := store.ReadAll()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	var child *task.Task
	for _, tt := range tasks {
		if tt.Title == "Child task" {
			child = tt
			break
		}
	}
	if child == nil {
		t.Fatal("child task not found")
	}
	if child.ParentID != parentID {
		t.Errorf("expected parent_id '%s', got '%s'", parentID, child.ParentID)
	}
}

func TestCreateValidation(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	// Test invalid type
	createType = "invalid"
	createPriority = 2
	createDescription = ""
	createParent = ""

	err := runCreate(createCmd, []string{"Test"})
	if err == nil || !strings.Contains(err.Error(), "invalid type") {
		t.Error("expected invalid type error")
	}

	// Test invalid priority
	createType = "feature"
	createPriority = 5

	err = runCreate(createCmd, []string{"Test"})
	if err == nil || !strings.Contains(err.Error(), "invalid priority") {
		t.Error("expected invalid priority error")
	}

	// Test invalid parent
	createType = "feature"
	createPriority = 2
	createParent = "atom-nonexistent"

	err = runCreate(createCmd, []string{"Test"})
	if err == nil || !strings.Contains(err.Error(), "parent task not found") {
		t.Error("expected parent not found error")
	}
}

func TestListTasks(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	// Create some tasks
	store := storage.New(dir)
	tasks := []*task.Task{
		{ID: "atom-0001", Title: "Bug fix", Type: task.TypeBug, Priority: 1, Status: task.StatusOpen},
		{ID: "atom-0002", Title: "Feature", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen},
		{ID: "atom-0003", Title: "Closed", Type: task.TypeFeature, Priority: 3, Status: task.StatusClosed},
	}
	for _, tt := range tasks {
		store.Append(tt)
	}

	// Test list (excludes closed by default)
	listStatus = ""
	listType = ""
	listPriority = 0
	listParent = ""
	listJSON = true

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runList(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result []*task.Task
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 tasks (non-closed), got %d", len(result))
	}

	// Test filter by type
	listType = "bug"
	r, w, _ = os.Pipe()
	os.Stdout = w
	runList(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	buf.ReadFrom(r)
	json.Unmarshal(buf.Bytes(), &result)

	if len(result) != 1 || result[0].Type != task.TypeBug {
		t.Error("filter by type failed")
	}

	// Test filter by status (should include closed)
	listType = ""
	listStatus = "closed"
	r, w, _ = os.Pipe()
	os.Stdout = w
	runList(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	buf.ReadFrom(r)
	json.Unmarshal(buf.Bytes(), &result)

	if len(result) != 1 || result[0].Status != task.StatusClosed {
		t.Error("filter by status failed")
	}
}

func TestShowTask(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	// Create parent and child
	store := storage.New(dir)
	parent := &task.Task{ID: "atom-0001", Title: "Parent", Type: task.TypeFeature, Priority: 1, Status: task.StatusOpen, Description: "Parent desc"}
	child := &task.Task{ID: "atom-0002", Title: "Child", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, ParentID: "atom-0001"}
	store.Append(parent)
	store.Append(child)

	// Test show with JSON
	showJSON = true

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runShow(showCmd, []string{"atom-0001"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("show failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result struct {
		ID          string        `json:"id"`
		Title       string        `json:"title"`
		Description string        `json:"description"`
		Children    []*task.Task  `json:"children"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result.ID != "atom-0001" {
		t.Errorf("expected id 'atom-0001', got '%s'", result.ID)
	}
	if result.Description != "Parent desc" {
		t.Errorf("expected description 'Parent desc', got '%s'", result.Description)
	}
	if len(result.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(result.Children))
	}
	if result.Children[0].ID != "atom-0002" {
		t.Errorf("expected child id 'atom-0002', got '%s'", result.Children[0].ID)
	}
}

func TestShowNotFound(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	showJSON = false
	err := runShow(showCmd, []string{"atom-nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Error("expected task not found error")
	}
}
