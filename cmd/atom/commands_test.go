package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scottwater/atoms/internal/merge"
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

func TestUpdateTask(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	// Create a task
	store := storage.New(dir)
	original := &task.Task{
		ID:          "atom-0001",
		Title:       "Original",
		Type:        task.TypeFeature,
		Priority:    2,
		Status:      task.StatusOpen,
		Description: "Original desc",
	}
	store.Append(original)

	// Test update status
	updateStatus = "in_progress"
	updatePriority = 0
	updateDescription = ""
	updateTitle = ""

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runUpdate(updateCmd, []string{"atom-0001"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]string
	json.Unmarshal(buf.Bytes(), &result)
	if result["status"] != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", result["status"])
	}

	// Verify in storage
	updated, _ := store.FindByID("atom-0001")
	if updated.Status != task.StatusInProgress {
		t.Errorf("expected status in_progress in storage, got %s", updated.Status)
	}

	// Test update priority
	updateStatus = ""
	updatePriority = 1
	r, w, _ = os.Pipe()
	os.Stdout = w
	runUpdate(updateCmd, []string{"atom-0001"})
	w.Close()
	os.Stdout = oldStdout

	updated, _ = store.FindByID("atom-0001")
	if updated.Priority != 1 {
		t.Errorf("expected priority 1, got %d", updated.Priority)
	}

	// Test update title
	updatePriority = 0
	updateTitle = "New Title"
	r, w, _ = os.Pipe()
	os.Stdout = w
	runUpdate(updateCmd, []string{"atom-0001"})
	w.Close()
	os.Stdout = oldStdout

	updated, _ = store.FindByID("atom-0001")
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", updated.Title)
	}
}

func TestUpdateValidation(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	store := storage.New(dir)
	store.Append(&task.Task{ID: "atom-0001", Title: "Test", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen})

	// Test invalid status
	updateStatus = "invalid"
	updatePriority = 0
	updateDescription = ""
	updateTitle = ""

	err := runUpdate(updateCmd, []string{"atom-0001"})
	if err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Error("expected invalid status error")
	}

	// Test invalid priority
	updateStatus = ""
	updatePriority = 5

	err = runUpdate(updateCmd, []string{"atom-0001"})
	if err == nil || !strings.Contains(err.Error(), "invalid priority") {
		t.Error("expected invalid priority error")
	}

	// Test no updates
	updatePriority = 0
	err = runUpdate(updateCmd, []string{"atom-0001"})
	if err == nil || !strings.Contains(err.Error(), "no updates specified") {
		t.Error("expected no updates error")
	}

	// Test task not found
	updateStatus = "open"
	err = runUpdate(updateCmd, []string{"atom-nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Error("expected task not found error")
	}
}

func TestCloseTask(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	store := storage.New(dir)
	store.Append(&task.Task{ID: "atom-0001", Title: "Test Task", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen})

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runClose(closeCmd, []string{"atom-0001"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("close failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Closed atom-0001") {
		t.Errorf("expected close confirmation, got '%s'", output)
	}

	// Verify in storage
	closed, _ := store.FindByID("atom-0001")
	if closed.Status != task.StatusClosed {
		t.Errorf("expected status closed, got %s", closed.Status)
	}
}

func TestCloseNotFound(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	err := runClose(closeCmd, []string{"atom-nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Error("expected task not found error")
	}
}

func TestReadyTasks(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	store := storage.New(dir)
	tasks := []*task.Task{
		{ID: "atom-0001", Title: "Open task", Type: task.TypeFeature, Priority: 1, Status: task.StatusOpen},
		{ID: "atom-0002", Title: "In progress", Type: task.TypeFeature, Priority: 2, Status: task.StatusInProgress},
		{ID: "atom-0003", Title: "Blocked", Type: task.TypeBug, Priority: 1, Status: task.StatusBlocked},
		{ID: "atom-0004", Title: "Closed", Type: task.TypeFeature, Priority: 3, Status: task.StatusClosed},
		{ID: "atom-0005", Title: "Another open", Type: task.TypeBug, Priority: 2, Status: task.StatusOpen},
	}
	for _, tt := range tasks {
		store.Append(tt)
	}

	readyJSON = true

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runReady(readyCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("ready failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result []*task.Task
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 ready tasks, got %d", len(result))
	}

	// Verify only open tasks are returned
	for _, rt := range result {
		if rt.Status != task.StatusOpen {
			t.Errorf("expected only open tasks, got %s with status %s", rt.ID, rt.Status)
		}
	}
}

func TestReadyEmpty(t *testing.T) {
	dir := setupTestDir(t)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	initPrefix = "atom"
	initQuiet = true
	runInit(initCmd, []string{})

	// Create only non-open tasks
	store := storage.New(dir)
	store.Append(&task.Task{ID: "atom-0001", Title: "In progress", Type: task.TypeFeature, Priority: 1, Status: task.StatusInProgress})
	store.Append(&task.Task{ID: "atom-0002", Title: "Closed", Type: task.TypeFeature, Priority: 2, Status: task.StatusClosed})

	readyJSON = false

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runReady(readyCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("ready failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No tasks ready") {
		t.Errorf("expected 'No tasks ready' message, got '%s'", output)
	}
}

func TestMergeCommand(t *testing.T) {
	dir := setupTestDir(t)

	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)
	oursTime := base.Add(time.Hour)
	theirsTime := base.Add(2 * time.Hour)

	// Create ancestor file
	ancestorPath := filepath.Join(dir, "ancestor.jsonl")
	ancestorStore := storage.NewWithPath(ancestorPath)
	ancestorStore.Create()
	ancestorStore.Append(&task.Task{
		ID: "atom-001", Title: "Original", Type: task.TypeFeature,
		Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base,
	})

	// Create ours file
	oursPath := filepath.Join(dir, "ours.jsonl")
	oursStore := storage.NewWithPath(oursPath)
	oursStore.Create()
	oursStore.Append(&task.Task{
		ID: "atom-001", Title: "Our Title", Type: task.TypeFeature,
		Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: oursTime,
	})
	oursStore.Append(&task.Task{
		ID: "atom-002", Title: "New in ours", Type: task.TypeBug,
		Priority: 1, Status: task.StatusOpen, CreatedAt: oursTime, UpdatedAt: oursTime,
	})

	// Create theirs file
	theirsPath := filepath.Join(dir, "theirs.jsonl")
	theirsStore := storage.NewWithPath(theirsPath)
	theirsStore.Create()
	theirsStore.Append(&task.Task{
		ID: "atom-001", Title: "Their Title", Type: task.TypeFeature,
		Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: theirsTime,
	})
	theirsStore.Append(&task.Task{
		ID: "atom-003", Title: "New in theirs", Type: task.TypeFeature,
		Priority: 3, Status: task.StatusOpen, CreatedAt: theirsTime, UpdatedAt: theirsTime,
	})

	// Run merge (note: runMerge calls os.Exit, so we test via the merge package directly)
	ancestor, _ := ancestorStore.ReadAll()
	ours, _ := oursStore.ReadAll()
	theirs, _ := theirsStore.ReadAll()

	merged := merge.Merge3Way(ancestor, ours, theirs)

	if len(merged) != 3 {
		t.Fatalf("expected 3 tasks after merge, got %d", len(merged))
	}

	// Find the original task and verify their title won (later timestamp)
	var originalTask *task.Task
	for _, tt := range merged {
		if tt.ID == "atom-001" {
			originalTask = tt
			break
		}
	}
	if originalTask == nil {
		t.Fatal("atom-001 not found in merged result")
	}
	if originalTask.Title != "Their Title" {
		t.Errorf("expected 'Their Title', got '%s'", originalTask.Title)
	}
}

func TestHelpCommand(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runHelp(helpCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify key sections are present
	if !strings.Contains(output, "atoms - Minimal git-backed task tracker") {
		t.Error("expected header in help output")
	}
	if !strings.Contains(output, "SETUP COMMANDS:") {
		t.Error("expected SETUP COMMANDS section")
	}
	if !strings.Contains(output, "TASK COMMANDS:") {
		t.Error("expected TASK COMMANDS section")
	}
	if !strings.Contains(output, "EXAMPLES:") {
		t.Error("expected EXAMPLES section")
	}
	if !strings.Contains(output, "WORKFLOW:") {
		t.Error("expected WORKFLOW section")
	}
	if !strings.Contains(output, "atom init") {
		t.Error("expected init command in help")
	}
	if !strings.Contains(output, "atom create") {
		t.Error("expected create command in help")
	}
	if !strings.Contains(output, "atom ready") {
		t.Error("expected ready command in help")
	}
}

func TestOnboardCommand(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runOnboard(onboardCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify key content is present
	if !strings.Contains(output, "--- BEGIN ATOM.MD CONTENT ---") {
		t.Error("expected BEGIN marker in onboard output")
	}
	if !strings.Contains(output, "--- END ATOM.MD CONTENT ---") {
		t.Error("expected END marker in onboard output")
	}
	if !strings.Contains(output, "## Task Tracking") {
		t.Error("expected Task Tracking header")
	}
	if !strings.Contains(output, "`atom ready`") {
		t.Error("expected atom ready command reference")
	}
	if !strings.Contains(output, "**Workflow:**") {
		t.Error("expected Workflow section")
	}
	if !strings.Contains(output, ".atoms.jsonl") {
		t.Error("expected .atoms.jsonl reference")
	}
}
