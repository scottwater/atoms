package merge

import (
	"testing"
	"time"

	"github.com/scottwater/atoms/internal/task"
)

func TestMerge3Way_NoConflicts(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].Title != "Task 1" {
		t.Errorf("expected title 'Task 1', got '%s'", result[0].Title)
	}
}

func TestMerge3Way_NewTaskOurs(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	ancestor := []*task.Task{}

	ours := []*task.Task{
		{ID: "atom-001", Title: "New task", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	theirs := []*task.Task{}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].ID != "atom-001" {
		t.Errorf("expected id 'atom-001', got '%s'", result[0].ID)
	}
}

func TestMerge3Way_NewTaskTheirs(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	ancestor := []*task.Task{}

	ours := []*task.Task{}

	theirs := []*task.Task{
		{ID: "atom-002", Title: "Their task", Type: task.TypeBug, Priority: 1, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].ID != "atom-002" {
		t.Errorf("expected id 'atom-002', got '%s'", result[0].ID)
	}
}

func TestMerge3Way_NewTaskBothBranches(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	ancestor := []*task.Task{}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Our task", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	theirs := []*task.Task{
		{ID: "atom-002", Title: "Their task", Type: task.TypeBug, Priority: 1, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result))
	}
}

func TestMerge3Way_DeletedByTheirs(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	theirs := []*task.Task{} // Deleted

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 0 {
		t.Fatalf("expected 0 tasks (deletion wins), got %d", len(result))
	}
}

func TestMerge3Way_DeletedByOurs(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{} // Deleted

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 0 {
		t.Fatalf("expected 0 tasks (deletion wins), got %d", len(result))
	}
}

func TestMerge3Way_TitleConflict_TimestampWins(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)
	oursTime := base.Add(time.Hour)
	theirsTime := base.Add(2 * time.Hour)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Original", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Our Title", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: oursTime},
	}

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Their Title", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: theirsTime},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].Title != "Their Title" {
		t.Errorf("expected 'Their Title' (later timestamp), got '%s'", result[0].Title)
	}
}

func TestMerge3Way_PriorityConflict_HigherWins(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)
	later := base.Add(time.Hour)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Task", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Task", Type: task.TypeFeature, Priority: 3, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: later},
	}

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Task", Type: task.TypeFeature, Priority: 1, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: later},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].Priority != 1 {
		t.Errorf("expected priority 1 (higher priority wins), got %d", result[0].Priority)
	}
}

func TestMerge3Way_StatusConflict_TimestampWins(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)
	oursTime := base.Add(2 * time.Hour)
	theirsTime := base.Add(time.Hour)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Task", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Task", Type: task.TypeFeature, Priority: 2, Status: task.StatusClosed, CreatedAt: base, UpdatedAt: oursTime},
	}

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Task", Type: task.TypeFeature, Priority: 2, Status: task.StatusInProgress, CreatedAt: base, UpdatedAt: theirsTime},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].Status != task.StatusClosed {
		t.Errorf("expected status 'closed' (later timestamp), got '%s'", result[0].Status)
	}
}

func TestMerge3Way_OneSideChanged(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)
	later := base.Add(time.Hour)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Original", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Original", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Updated Title", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: later},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].Title != "Updated Title" {
		t.Errorf("expected 'Updated Title' (their change), got '%s'", result[0].Title)
	}
}

func TestMerge3Way_BothChangedDifferentFields(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)
	oursTime := base.Add(time.Hour)
	theirsTime := base.Add(30 * time.Minute)

	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Original", Description: "Orig desc", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "New Title", Description: "Orig desc", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: oursTime},
	}

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Original", Description: "New desc", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, UpdatedAt: theirsTime},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", result[0].Title)
	}
	if result[0].Description != "New desc" {
		t.Errorf("expected description 'New desc', got '%s'", result[0].Description)
	}
}

func TestMerge3Way_CompositeKeyMatching(t *testing.T) {
	base := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)
	later := base.Add(time.Hour)

	// Same ID but different created_at means different tasks
	ancestor := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, CreatedBy: "alice", UpdatedAt: base},
	}

	ours := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, CreatedBy: "alice", UpdatedAt: base},
		{ID: "atom-001", Title: "Same ID different time", Type: task.TypeBug, Priority: 1, Status: task.StatusOpen, CreatedAt: later, CreatedBy: "bob", UpdatedAt: later},
	}

	theirs := []*task.Task{
		{ID: "atom-001", Title: "Task 1", Type: task.TypeFeature, Priority: 2, Status: task.StatusOpen, CreatedAt: base, CreatedBy: "alice", UpdatedAt: base},
	}

	result := Merge3Way(ancestor, ours, theirs)

	if len(result) != 2 {
		t.Fatalf("expected 2 tasks (different composite keys), got %d", len(result))
	}
}
