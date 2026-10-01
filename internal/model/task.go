package model

import "time"

// Status is the lifecycle state of a Task.
type Status string

const (
	StatusTodo       Status = "TODO"
	StatusInProgress Status = "IN_PROGRESS"
	StatusDone       Status = "DONE"
)

// Valid reports whether s is one of the known task statuses.
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// Task is a unit of work that belongs to a Project and may be assigned to a User.
type Task struct {
	ID          int64     `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Status      Status    `db:"status" json:"status"`
	ProjectID   int64     `db:"project_id" json:"projectId"`
	AssigneeID  *int64    `db:"assignee_id" json:"assigneeId,omitempty"`
	DueDate     *Date     `db:"due_date" json:"dueDate,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}
