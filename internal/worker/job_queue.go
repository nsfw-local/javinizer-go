package worker

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// FileResult represents the result of processing a single file
type FileResult struct {
	FilePath  string      `json:"file_path"`
	MovieID   string      `json:"movie_id"`
	Status    JobStatus   `json:"status"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	StartedAt time.Time   `json:"started_at"`
	EndedAt   *time.Time  `json:"ended_at,omitempty"`
}

// BatchJob represents a batch processing job
type BatchJob struct {
	ID          string                 `json:"id"`
	Status      JobStatus              `json:"status"`
	TotalFiles  int                    `json:"total_files"`
	Completed   int                    `json:"completed"`
	Failed      int                    `json:"failed"`
	Files       []string               `json:"files"`
	Results     map[string]*FileResult `json:"results"` // keyed by file path
	Progress    float64                `json:"progress"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CancelFunc  context.CancelFunc     `json:"-"`
	mu          sync.RWMutex           `json:"-"`
}

// JobQueue manages batch jobs
type JobQueue struct {
	jobs map[string]*BatchJob
	mu   sync.RWMutex
}

// NewJobQueue creates a new job queue
func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs: make(map[string]*BatchJob),
	}
}

// CreateJob creates a new batch job
func (jq *JobQueue) CreateJob(files []string) *BatchJob {
	job := &BatchJob{
		ID:         uuid.New().String(),
		Status:     JobStatusPending,
		TotalFiles: len(files),
		Files:      files,
		Results:    make(map[string]*FileResult),
		StartedAt:  time.Now(),
	}

	jq.mu.Lock()
	jq.jobs[job.ID] = job
	jq.mu.Unlock()

	return job
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(id string) (*BatchJob, bool) {
	jq.mu.RLock()
	defer jq.mu.RUnlock()
	job, ok := jq.jobs[id]
	return job, ok
}

// DeleteJob removes a job from the queue
func (jq *JobQueue) DeleteJob(id string) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	delete(jq.jobs, id)
}

// ListJobs returns all jobs
func (jq *JobQueue) ListJobs() []*BatchJob {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	jobs := make([]*BatchJob, 0, len(jq.jobs))
	for _, job := range jq.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// UpdateFileResult updates the result for a specific file in the job
func (job *BatchJob) UpdateFileResult(filePath string, result *FileResult) {
	job.mu.Lock()
	defer job.mu.Unlock()

	job.Results[filePath] = result

	// Update counters
	completed := 0
	failed := 0
	for _, r := range job.Results {
		if r.Status == JobStatusCompleted {
			completed++
		} else if r.Status == JobStatusFailed {
			failed++
		}
	}
	job.Completed = completed
	job.Failed = failed
	job.Progress = float64(completed+failed) / float64(job.TotalFiles) * 100
}

// MarkStarted marks the job as started
func (job *BatchJob) MarkStarted() {
	job.mu.Lock()
	defer job.mu.Unlock()
	job.Status = JobStatusRunning
	job.StartedAt = time.Now()
}

// MarkCompleted marks the job as completed
func (job *BatchJob) MarkCompleted() {
	job.mu.Lock()
	defer job.mu.Unlock()
	job.Status = JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now
	job.Progress = 100
}

// MarkFailed marks the job as failed
func (job *BatchJob) MarkFailed() {
	job.mu.Lock()
	defer job.mu.Unlock()
	job.Status = JobStatusFailed
	now := time.Now()
	job.CompletedAt = &now
}

// MarkCancelled marks the job as cancelled
func (job *BatchJob) MarkCancelled() {
	job.mu.Lock()
	defer job.mu.Unlock()
	job.Status = JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now
}

// Cancel cancels the job
func (job *BatchJob) Cancel() {
	if job.CancelFunc != nil {
		job.CancelFunc()
	}
	job.MarkCancelled()
}

// GetStatus returns a thread-safe copy of the job status
func (job *BatchJob) GetStatus() *BatchJob {
	job.mu.RLock()
	defer job.mu.RUnlock()

	// Create a copy to avoid race conditions
	results := make(map[string]*FileResult)
	for k, v := range job.Results {
		results[k] = v
	}

	completedAt := job.CompletedAt
	if completedAt != nil {
		t := *completedAt
		completedAt = &t
	}

	return &BatchJob{
		ID:          job.ID,
		Status:      job.Status,
		TotalFiles:  job.TotalFiles,
		Completed:   job.Completed,
		Failed:      job.Failed,
		Files:       job.Files,
		Results:     results,
		Progress:    job.Progress,
		StartedAt:   job.StartedAt,
		CompletedAt: completedAt,
	}
}
