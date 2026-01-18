package task

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	task := New("Test task", TypeFeature, 2)

	if task.Title != "Test task" {
		t.Errorf("expected title 'Test task', got '%s'", task.Title)
	}
	if task.Type != TypeFeature {
		t.Errorf("expected type feature, got %s", task.Type)
	}
	if task.Priority != 2 {
		t.Errorf("expected priority 2, got %d", task.Priority)
	}
	if task.Status != StatusOpen {
		t.Errorf("expected status open, got %s", task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestClose(t *testing.T) {
	task := New("Test task", TypeBug, 1)
	originalUpdated := task.UpdatedAt

	time.Sleep(time.Millisecond)
	task.Close()

	if task.Status != StatusClosed {
		t.Errorf("expected status closed, got %s", task.Status)
	}
	if !task.UpdatedAt.After(originalUpdated) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestCompositeKey(t *testing.T) {
	task := &Task{
		ID:        "atom-a3f2",
		CreatedAt: time.Date(2025, 1, 18, 10, 30, 0, 0, time.UTC),
		CreatedBy: "scott",
	}

	key := task.CompositeKey()
	expected := "atom-a3f2|2025-01-18T10:30:00Z|scott"

	if key != expected {
		t.Errorf("expected key '%s', got '%s'", expected, key)
	}
}
