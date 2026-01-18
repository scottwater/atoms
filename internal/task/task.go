package task

import "time"

type Type string

const (
	TypeFeature Type = "feature"
	TypeBug     Type = "bug"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusClosed     Status = "closed"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Type        Type      `json:"type"`
	Priority    int       `json:"priority"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	ParentID    string    `json:"parent_id,omitempty"`
}

func New(title string, taskType Type, priority int) *Task {
	now := time.Now().UTC()
	return &Task{
		Title:     title,
		Type:      taskType,
		Priority:  priority,
		Status:    StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Task) Close() {
	t.Status = StatusClosed
	t.UpdatedAt = time.Now().UTC()
}

func (t *Task) Update() {
	t.UpdatedAt = time.Now().UTC()
}

func (t *Task) CompositeKey() string {
	return t.ID + "|" + t.CreatedAt.Format(time.RFC3339) + "|" + t.CreatedBy
}
