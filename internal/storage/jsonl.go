package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/scottwater/atoms/internal/task"
)

const DefaultFilename = ".atoms.jsonl"

type Storage struct {
	path string
}

func New(dir string) *Storage {
	return &Storage{
		path: filepath.Join(dir, DefaultFilename),
	}
}

func NewWithPath(path string) *Storage {
	return &Storage{path: path}
}

func (s *Storage) Path() string {
	return s.path
}

func (s *Storage) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

func (s *Storage) Create() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	return f.Close()
}

func (s *Storage) ReadAll() ([]*task.Task, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var tasks []*task.Task
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var t task.Task
		if err := json.Unmarshal(line, &t); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		tasks = append(tasks, &t)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return tasks, nil
}

func (s *Storage) Append(t *task.Task) error {
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for append: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write task: %w", err)
	}

	return nil
}

func (s *Storage) WriteAll(tasks []*task.Task) error {
	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	for _, t := range tasks {
		data, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("failed to marshal task %s: %w", t.ID, err)
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write task %s: %w", t.ID, err)
		}
	}

	return nil
}

func (s *Storage) FindByID(id string) (*task.Task, error) {
	tasks, err := s.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, nil
}

func (s *Storage) Update(updated *task.Task) error {
	tasks, err := s.ReadAll()
	if err != nil {
		return err
	}

	found := false
	for i, t := range tasks {
		if t.ID == updated.ID {
			tasks[i] = updated
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task %s not found", updated.ID)
	}

	return s.WriteAll(tasks)
}
