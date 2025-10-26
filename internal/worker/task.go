package worker

import (
	"context"
	"time"
)

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeScan     TaskType = "scan"
	TaskTypeScrape   TaskType = "scrape"
	TaskTypeDownload TaskType = "download"
	TaskTypeOrganize TaskType = "organize"
	TaskTypeNFO      TaskType = "nfo"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusSuccess  TaskStatus = "success"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusCanceled TaskStatus = "canceled"
)

// Task represents a unit of work to be executed
type Task interface {
	// ID returns a unique identifier for this task
	ID() string

	// Type returns the type of task
	Type() TaskType

	// Execute runs the task and returns an error if it fails
	Execute(ctx context.Context) error

	// Description returns a human-readable description
	Description() string
}

// TaskResult holds the result of a completed task
type TaskResult struct {
	TaskID      string
	Type        TaskType
	Status      TaskStatus
	Error       error
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	BytesTotal  int64
	BytesDone   int64
	Description string
}

// BaseTask provides common task functionality
type BaseTask struct {
	id          string
	taskType    TaskType
	description string
}

func (t *BaseTask) ID() string {
	return t.id
}

func (t *BaseTask) Type() TaskType {
	return t.taskType
}

func (t *BaseTask) Description() string {
	return t.description
}
