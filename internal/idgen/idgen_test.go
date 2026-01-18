package idgen

import (
	"strings"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	gen := New("atom")
	now := time.Now()

	id, err := gen.Generate("Test task", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(id, "atom-") {
		t.Errorf("expected id to start with 'atom-', got '%s'", id)
	}

	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		t.Fatalf("expected id format 'prefix-hash', got '%s'", id)
	}

	hash := parts[1]
	if len(hash) < MinHashLength || len(hash) > MaxHashLength+2 {
		t.Errorf("expected hash length between %d and %d, got %d", MinHashLength, MaxHashLength+2, len(hash))
	}
}

func TestGenerateUnique(t *testing.T) {
	gen := New("atom")
	now := time.Now()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := gen.Generate("Test task", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ids[id] {
			t.Errorf("duplicate id generated: %s", id)
		}
		ids[id] = true
	}
}

func TestGenerateCustomPrefix(t *testing.T) {
	gen := New("task")
	now := time.Now()

	id, err := gen.Generate("Test", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(id, "task-") {
		t.Errorf("expected id to start with 'task-', got '%s'", id)
	}
}

func TestSetExisting(t *testing.T) {
	gen := New("atom")
	gen.SetExisting([]string{"atom-a3f2", "atom-b7c4"})

	if !gen.existing["atom-a3f2"] {
		t.Error("expected atom-a3f2 to be in existing set")
	}
	if !gen.existing["atom-b7c4"] {
		t.Error("expected atom-b7c4 to be in existing set")
	}
}
