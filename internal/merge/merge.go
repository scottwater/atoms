package merge

import (
	"time"

	"github.com/scottwater/atoms/internal/task"
)

// Merge3Way performs a 3-way merge of task lists.
// ancestor is the common base, ours is our changes, theirs is their changes.
// Returns the merged task list.
func Merge3Way(ancestor, ours, theirs []*task.Task) []*task.Task {
	ancestorMap := taskMapByKey(ancestor)
	oursMap := taskMapByKey(ours)
	theirsMap := taskMapByKey(theirs)

	seen := make(map[string]bool)
	var result []*task.Task

	// Process tasks from ours
	for key, ourTask := range oursMap {
		seen[key] = true
		ancestorTask := ancestorMap[key]
		theirTask := theirsMap[key]

		if theirTask == nil {
			// Task exists in ours but not theirs
			if ancestorTask != nil {
				// Was in ancestor, deleted by theirs - deletion wins
				continue
			}
			// New in ours, keep it
			result = append(result, ourTask)
		} else {
			// Task exists in both ours and theirs
			merged := mergeTask(ancestorTask, ourTask, theirTask)
			result = append(result, merged)
		}
	}

	// Process tasks only in theirs (not in ours)
	for key, theirTask := range theirsMap {
		if seen[key] {
			continue
		}
		ancestorTask := ancestorMap[key]

		if ancestorTask != nil {
			// Was in ancestor, deleted by ours - deletion wins
			continue
		}
		// New in theirs, keep it
		result = append(result, theirTask)
	}

	return result
}

// taskMapByKey creates a map of tasks keyed by composite key
func taskMapByKey(tasks []*task.Task) map[string]*task.Task {
	m := make(map[string]*task.Task)
	for _, t := range tasks {
		m[t.CompositeKey()] = t
	}
	return m
}

// mergeTask merges a single task that exists in both branches
func mergeTask(ancestor, ours, theirs *task.Task) *task.Task {
	if ancestor == nil {
		// Both added the same task (same composite key) - use timestamp to resolve
		if theirs.UpdatedAt.After(ours.UpdatedAt) {
			return theirs
		}
		return ours
	}

	// Start with ours as base
	merged := &task.Task{
		ID:        ours.ID,
		CreatedAt: ours.CreatedAt,
		CreatedBy: ours.CreatedBy,
	}

	// Merge each field
	merged.Title = mergeStringField(ancestor.Title, ours.Title, theirs.Title, ours.UpdatedAt, theirs.UpdatedAt)
	merged.Description = mergeStringField(ancestor.Description, ours.Description, theirs.Description, ours.UpdatedAt, theirs.UpdatedAt)
	merged.Type = mergeTypeField(ancestor.Type, ours.Type, theirs.Type, ours.UpdatedAt, theirs.UpdatedAt)
	merged.Priority = mergePriorityField(ancestor.Priority, ours.Priority, theirs.Priority)
	merged.Status = mergeStatusField(ancestor.Status, ours.Status, theirs.Status, ours.UpdatedAt, theirs.UpdatedAt)
	merged.ParentID = mergeStringField(ancestor.ParentID, ours.ParentID, theirs.ParentID, ours.UpdatedAt, theirs.UpdatedAt)

	// UpdatedAt is the later of the two
	if theirs.UpdatedAt.After(ours.UpdatedAt) {
		merged.UpdatedAt = theirs.UpdatedAt
	} else {
		merged.UpdatedAt = ours.UpdatedAt
	}

	return merged
}

// mergeStringField merges a string field using timestamp resolution
func mergeStringField(ancestor, ours, theirs string, oursTime, theirsTime time.Time) string {
	// No conflict if values are the same
	if ours == theirs {
		return ours
	}

	// If only one side changed from ancestor
	if ours == ancestor && theirs != ancestor {
		return theirs
	}
	if theirs == ancestor && ours != ancestor {
		return ours
	}

	// Both changed - later timestamp wins
	if theirsTime.After(oursTime) {
		return theirs
	}
	return ours
}

// mergePriorityField merges priority - higher priority (lower number) wins
func mergePriorityField(ancestor, ours, theirs int) int {
	// No conflict if values are the same
	if ours == theirs {
		return ours
	}

	// If only one side changed from ancestor
	if ours == ancestor && theirs != ancestor {
		return theirs
	}
	if theirs == ancestor && ours != ancestor {
		return ours
	}

	// Both changed - higher priority (lower number) wins
	if theirs < ours {
		return theirs
	}
	return ours
}

// mergeTypeField merges type using timestamp resolution
func mergeTypeField(ancestor, ours, theirs task.Type, oursTime, theirsTime time.Time) task.Type {
	if ours == theirs {
		return ours
	}
	if ours == ancestor && theirs != ancestor {
		return theirs
	}
	if theirs == ancestor && ours != ancestor {
		return ours
	}
	if theirsTime.After(oursTime) {
		return theirs
	}
	return ours
}

// mergeStatusField merges status using timestamp resolution
func mergeStatusField(ancestor, ours, theirs task.Status, oursTime, theirsTime time.Time) task.Status {
	if ours == theirs {
		return ours
	}
	if ours == ancestor && theirs != ancestor {
		return theirs
	}
	if theirs == ancestor && ours != ancestor {
		return ours
	}
	if theirsTime.After(oursTime) {
		return theirs
	}
	return ours
}
