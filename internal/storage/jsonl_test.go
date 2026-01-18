package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scottwater/atoms/internal/task"
)

func TestStorageCreateAndExists(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	if s.Exists() {
		t.Error("expected file to not exist initially")
	}

	if err := s.Create(); err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	if !s.Exists() {
		t.Error("expected file to exist after create")
	}
}

func TestStorageAppendAndReadAll(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	if err := s.Create(); err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	now := time.Now().UTC()
	task1 := &task.Task{
		ID:        "atom-a3f2",
		Title:     "First task",
		Type:      task.TypeFeature,
		Priority:  1,
		Status:    task.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.Append(task1); err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	task2 := &task.Task{
		ID:        "atom-b7c4",
		Title:     "Second task",
		Type:      task.TypeBug,
		Priority:  2,
		Status:    task.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.Append(task2); err != nil {
		t.Fatalf("failed to append second task: %v", err)
	}

	tasks, err := s.ReadAll()
	if err != nil {
		t.Fatalf("failed to read all: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	if tasks[0].ID != "atom-a3f2" {
		t.Errorf("expected first task id 'atom-a3f2', got '%s'", tasks[0].ID)
	}
	if tasks[1].ID != "atom-b7c4" {
		t.Errorf("expected second task id 'atom-b7c4', got '%s'", tasks[1].ID)
	}
}

func TestStorageFindByID(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.Create()

	now := time.Now().UTC()
	task1 := &task.Task{
		ID:        "atom-find",
		Title:     "Find me",
		Type:      task.TypeFeature,
		Priority:  1,
		Status:    task.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.Append(task1)

	found, err := s.FindByID("atom-find")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find task")
	}
	if found.Title != "Find me" {
		t.Errorf("expected title 'Find me', got '%s'", found.Title)
	}

	notFound, err := s.FindByID("atom-nope")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent task")
	}
}

func TestStorageUpdate(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.Create()

	now := time.Now().UTC()
	original := &task.Task{
		ID:        "atom-upd",
		Title:     "Original title",
		Type:      task.TypeFeature,
		Priority:  2,
		Status:    task.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.Append(original)

	updated := &task.Task{
		ID:        "atom-upd",
		Title:     "Updated title",
		Type:      task.TypeFeature,
		Priority:  1,
		Status:    task.StatusInProgress,
		CreatedAt: now,
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.Update(updated); err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	found, _ := s.FindByID("atom-upd")
	if found.Title != "Updated title" {
		t.Errorf("expected title 'Updated title', got '%s'", found.Title)
	}
	if found.Priority != 1 {
		t.Errorf("expected priority 1, got %d", found.Priority)
	}
	if found.Status != task.StatusInProgress {
		t.Errorf("expected status in_progress, got %s", found.Status)
	}
}

func TestStorageWriteAll(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.Create()

	now := time.Now().UTC()
	tasks := []*task.Task{
		{ID: "atom-1", Title: "Task 1", Type: task.TypeFeature, Priority: 1, Status: task.StatusOpen, CreatedAt: now, UpdatedAt: now},
		{ID: "atom-2", Title: "Task 2", Type: task.TypeBug, Priority: 2, Status: task.StatusOpen, CreatedAt: now, UpdatedAt: now},
	}

	if err := s.WriteAll(tasks); err != nil {
		t.Fatalf("failed to write all: %v", err)
	}

	read, err := s.ReadAll()
	if err != nil {
		t.Fatalf("failed to read all: %v", err)
	}

	if len(read) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(read))
	}
}

func TestStoragePath(t *testing.T) {
	dir := "/tmp/test"
	s := New(dir)

	expected := filepath.Join(dir, DefaultFilename)
	if s.Path() != expected {
		t.Errorf("expected path '%s', got '%s'", expected, s.Path())
	}
}

func TestStorageReadAllNonExistent(t *testing.T) {
	dir := t.TempDir()
	s := NewWithPath(filepath.Join(dir, "nonexistent.jsonl"))

	tasks, err := s.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks != nil {
		t.Errorf("expected nil tasks for non-existent file, got %v", tasks)
	}
}

func TestStorageReadEmptyLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".atoms.jsonl")

	content := `{"id":"atom-1","title":"Task","type":"feature","priority":1,"status":"open","created_at":"2025-01-18T10:00:00Z","updated_at":"2025-01-18T10:00:00Z"}

{"id":"atom-2","title":"Task 2","type":"bug","priority":2,"status":"open","created_at":"2025-01-18T10:00:00Z","updated_at":"2025-01-18T10:00:00Z"}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	s := NewWithPath(path)
	tasks, err := s.ReadAll()
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}
