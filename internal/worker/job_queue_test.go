package worker

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobQueue_CreateGetDeleteList(t *testing.T) {
	t.Run("Create and get job", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4", "file2.mkv", "file3.avi"}

		job := jq.CreateJob(files)

		// Verify job creation
		assert.NotEmpty(t, job.ID)
		assert.Equal(t, JobStatusPending, job.Status)
		assert.Equal(t, 3, job.TotalFiles)
		assert.Equal(t, files, job.Files)
		assert.NotNil(t, job.Results)
		assert.Empty(t, job.Results)
		assert.Equal(t, 0, job.Completed)
		assert.Equal(t, 0, job.Failed)
		assert.Equal(t, 0.0, job.Progress)
		assert.False(t, job.StartedAt.IsZero())
		assert.Nil(t, job.CompletedAt)

		// Retrieve the job
		retrieved, ok := jq.GetJob(job.ID)
		require.True(t, ok, "Job should exist")
		assert.Equal(t, job.ID, retrieved.ID)
		assert.Equal(t, job.TotalFiles, retrieved.TotalFiles)
	})

	t.Run("Get non-existent job", func(t *testing.T) {
		jq := NewJobQueue()

		retrieved, ok := jq.GetJob("non-existent-id")
		assert.False(t, ok, "Job should not exist")
		assert.Nil(t, retrieved)
	})

	t.Run("Delete job", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4"}

		job := jq.CreateJob(files)
		jobID := job.ID

		// Verify job exists
		_, ok := jq.GetJob(jobID)
		require.True(t, ok, "Job should exist before deletion")

		// Delete job
		jq.DeleteJob(jobID)

		// Verify job is deleted
		_, ok = jq.GetJob(jobID)
		assert.False(t, ok, "Job should not exist after deletion")
	})

	t.Run("List jobs", func(t *testing.T) {
		jq := NewJobQueue()

		// Initially empty
		jobs := jq.ListJobs()
		assert.Empty(t, jobs)

		// Create multiple jobs
		job1 := jq.CreateJob([]string{"file1.mp4"})
		job2 := jq.CreateJob([]string{"file2.mkv", "file3.avi"})
		job3 := jq.CreateJob([]string{"file4.mp4"})

		// List should contain all jobs
		jobs = jq.ListJobs()
		assert.Len(t, jobs, 3)

		// Verify all job IDs are present
		jobIDs := make(map[string]bool)
		for _, job := range jobs {
			jobIDs[job.ID] = true
		}
		assert.True(t, jobIDs[job1.ID], "Job1 should be in list")
		assert.True(t, jobIDs[job2.ID], "Job2 should be in list")
		assert.True(t, jobIDs[job3.ID], "Job3 should be in list")

		// Delete one job
		jq.DeleteJob(job2.ID)

		// List should have 2 jobs
		jobs = jq.ListJobs()
		assert.Len(t, jobs, 2)
	})

	t.Run("Empty files list", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{})

		assert.Equal(t, 0, job.TotalFiles)
		assert.Empty(t, job.Files)
	})
}

func TestBatchJob_UpdateFileResult(t *testing.T) {
	t.Run("Update single file result", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4", "file2.mkv", "file3.avi"}
		job := jq.CreateJob(files)

		now := time.Now()
		result := &FileResult{
			FilePath:  "file1.mp4",
			MovieID:   "IPX-123",
			Status:    JobStatusCompleted,
			StartedAt: now,
			EndedAt:   &now,
		}

		job.UpdateFileResult("file1.mp4", result)

		// Verify result is stored
		assert.Len(t, job.Results, 1)
		assert.Equal(t, result, job.Results["file1.mp4"])

		// Verify counters
		assert.Equal(t, 1, job.Completed)
		assert.Equal(t, 0, job.Failed)
		assert.InDelta(t, 33.33, job.Progress, 0.1) // 1/3 * 100
	})

	t.Run("Update multiple file results with mixed status", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4", "file2.mkv", "file3.avi", "file4.mp4"}
		job := jq.CreateJob(files)

		now := time.Now()

		// Complete first file
		job.UpdateFileResult("file1.mp4", &FileResult{
			FilePath:  "file1.mp4",
			Status:    JobStatusCompleted,
			StartedAt: now,
		})
		assert.Equal(t, 1, job.Completed)
		assert.Equal(t, 0, job.Failed)
		assert.Equal(t, 25.0, job.Progress)

		// Complete second file
		job.UpdateFileResult("file2.mkv", &FileResult{
			FilePath:  "file2.mkv",
			Status:    JobStatusCompleted,
			StartedAt: now,
		})
		assert.Equal(t, 2, job.Completed)
		assert.Equal(t, 0, job.Failed)
		assert.Equal(t, 50.0, job.Progress)

		// Fail third file
		job.UpdateFileResult("file3.avi", &FileResult{
			FilePath:  "file3.avi",
			Status:    JobStatusFailed,
			Error:     "scraper error",
			StartedAt: now,
		})
		assert.Equal(t, 2, job.Completed)
		assert.Equal(t, 1, job.Failed)
		assert.Equal(t, 75.0, job.Progress) // (2+1)/4 * 100

		// Complete fourth file
		job.UpdateFileResult("file4.mp4", &FileResult{
			FilePath:  "file4.mp4",
			Status:    JobStatusCompleted,
			StartedAt: now,
		})
		assert.Equal(t, 3, job.Completed)
		assert.Equal(t, 1, job.Failed)
		assert.Equal(t, 100.0, job.Progress)
	})

	t.Run("Update same file result multiple times", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4"}
		job := jq.CreateJob(files)

		now := time.Now()

		// Initially running
		job.UpdateFileResult("file1.mp4", &FileResult{
			FilePath:  "file1.mp4",
			Status:    JobStatusRunning,
			StartedAt: now,
		})
		assert.Equal(t, 0, job.Completed)
		assert.Equal(t, 0, job.Failed)
		assert.Equal(t, 0.0, job.Progress)

		// Then completed
		job.UpdateFileResult("file1.mp4", &FileResult{
			FilePath:  "file1.mp4",
			Status:    JobStatusCompleted,
			MovieID:   "IPX-123",
			StartedAt: now,
		})
		assert.Equal(t, 1, job.Completed)
		assert.Equal(t, 0, job.Failed)
		assert.Equal(t, 100.0, job.Progress)

		// Verify only one result exists
		assert.Len(t, job.Results, 1)
	})

	t.Run("Progress calculation with pending files", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4", "file2.mkv", "file3.avi"}
		job := jq.CreateJob(files)

		now := time.Now()

		// Only update 1 out of 3 files
		job.UpdateFileResult("file1.mp4", &FileResult{
			FilePath:  "file1.mp4",
			Status:    JobStatusCompleted,
			StartedAt: now,
		})

		// Progress should be 33.33% (1/3), not 100%
		assert.InDelta(t, 33.33, job.Progress, 0.1)
		assert.Equal(t, 1, job.Completed)
		assert.Equal(t, 2, job.TotalFiles-job.Completed-job.Failed) // 2 pending
	})
}

func TestBatchJob_StatusTransitions(t *testing.T) {
	t.Run("MarkStarted", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{"file1.mp4"})

		assert.Equal(t, JobStatusPending, job.Status)
		initialStartTime := job.StartedAt

		time.Sleep(10 * time.Millisecond) // Ensure time difference

		job.MarkStarted()

		assert.Equal(t, JobStatusRunning, job.Status)
		assert.True(t, job.StartedAt.After(initialStartTime), "StartedAt should be updated")
		assert.Nil(t, job.CompletedAt)
	})

	t.Run("MarkCompleted", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{"file1.mp4"})
		job.MarkStarted()

		beforeCompletion := time.Now()
		job.MarkCompleted()
		afterCompletion := time.Now()

		assert.Equal(t, JobStatusCompleted, job.Status)
		assert.Equal(t, 100.0, job.Progress)
		require.NotNil(t, job.CompletedAt)
		assert.True(t, job.CompletedAt.After(beforeCompletion) || job.CompletedAt.Equal(beforeCompletion))
		assert.True(t, job.CompletedAt.Before(afterCompletion) || job.CompletedAt.Equal(afterCompletion))
	})

	t.Run("MarkFailed", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{"file1.mp4"})
		job.MarkStarted()

		beforeFailure := time.Now()
		job.MarkFailed()
		afterFailure := time.Now()

		assert.Equal(t, JobStatusFailed, job.Status)
		require.NotNil(t, job.CompletedAt)
		assert.True(t, job.CompletedAt.After(beforeFailure) || job.CompletedAt.Equal(beforeFailure))
		assert.True(t, job.CompletedAt.Before(afterFailure) || job.CompletedAt.Equal(afterFailure))
	})

	t.Run("MarkCancelled", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{"file1.mp4"})
		job.MarkStarted()

		beforeCancellation := time.Now()
		job.MarkCancelled()
		afterCancellation := time.Now()

		assert.Equal(t, JobStatusCancelled, job.Status)
		require.NotNil(t, job.CompletedAt)
		assert.True(t, job.CompletedAt.After(beforeCancellation) || job.CompletedAt.Equal(beforeCancellation))
		assert.True(t, job.CompletedAt.Before(afterCancellation) || job.CompletedAt.Equal(afterCancellation))
	})

	t.Run("Full workflow: pending -> running -> completed", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4", "file2.mkv"}
		job := jq.CreateJob(files)

		// Start as pending
		assert.Equal(t, JobStatusPending, job.Status)

		// Mark as running
		job.MarkStarted()
		assert.Equal(t, JobStatusRunning, job.Status)
		assert.Nil(t, job.CompletedAt)

		// Process files
		now := time.Now()
		job.UpdateFileResult("file1.mp4", &FileResult{
			FilePath:  "file1.mp4",
			Status:    JobStatusCompleted,
			StartedAt: now,
		})
		job.UpdateFileResult("file2.mkv", &FileResult{
			FilePath:  "file2.mkv",
			Status:    JobStatusCompleted,
			StartedAt: now,
		})

		// Mark as completed
		job.MarkCompleted()
		assert.Equal(t, JobStatusCompleted, job.Status)
		assert.NotNil(t, job.CompletedAt)
		assert.Equal(t, 100.0, job.Progress)
	})
}

func TestBatchJob_GetStatus(t *testing.T) {
	t.Run("Returns copy with all fields", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4", "file2.mkv"}
		job := jq.CreateJob(files)
		job.MarkStarted()

		now := time.Now()
		job.UpdateFileResult("file1.mp4", &FileResult{
			FilePath:  "file1.mp4",
			Status:    JobStatusCompleted,
			StartedAt: now,
		})

		status := job.GetStatus()

		// Verify all fields are copied
		assert.Equal(t, job.ID, status.ID)
		assert.Equal(t, job.Status, status.Status)
		assert.Equal(t, job.TotalFiles, status.TotalFiles)
		assert.Equal(t, job.Completed, status.Completed)
		assert.Equal(t, job.Failed, status.Failed)
		assert.Equal(t, job.Files, status.Files)
		assert.Equal(t, job.Progress, status.Progress)
		assert.Equal(t, job.StartedAt, status.StartedAt)
		assert.Len(t, status.Results, 1)
	})

	t.Run("Deep copy of FileResults - map and FileResults are independent", func(t *testing.T) {
		jq := NewJobQueue()
		files := []string{"file1.mp4", "file2.mkv"}
		job := jq.CreateJob(files)

		now := time.Now()
		result1 := &FileResult{
			FilePath:  "file1.mp4",
			Status:    JobStatusCompleted,
			MovieID:   "IPX-123",
			StartedAt: now,
		}
		job.UpdateFileResult("file1.mp4", result1)

		// Get status copy
		status := job.GetStatus()

		// Verify FileResult objects are NOT shared (deep copy)
		// FileResults should be independent to prevent concurrent mutations
		assert.NotSame(t, job.Results["file1.mp4"], status.Results["file1.mp4"],
			"FileResult pointers should be different (deep copy)")

		// Verify fields are equal but independent
		assert.Equal(t, job.Results["file1.mp4"].MovieID, status.Results["file1.mp4"].MovieID,
			"FileResult fields should be equal")

		// Modifying a FileResult in the copy should NOT affect original
		status.Results["file1.mp4"].MovieID = "MODIFIED-999"
		assert.Equal(t, "IPX-123", job.Results["file1.mp4"].MovieID,
			"Original FileResult should remain unchanged (deep copy)")
		assert.Equal(t, "MODIFIED-999", status.Results["file1.mp4"].MovieID,
			"Copy FileResult should be modified")

		// Adding new entries to the copy's map doesn't affect original
		status.Results["file2.mkv"] = &FileResult{
			FilePath:  "file2.mkv",
			Status:    JobStatusCompleted,
			StartedAt: now,
		}
		assert.Len(t, status.Results, 2, "Copy should have 2 results")
		assert.Len(t, job.Results, 1, "Original should still have 1 result (map is independent)")
	})

	t.Run("Copies CompletedAt correctly when nil", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{"file1.mp4"})

		status := job.GetStatus()

		assert.Nil(t, job.CompletedAt)
		assert.Nil(t, status.CompletedAt)
	})

	t.Run("Copies CompletedAt correctly when set", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{"file1.mp4"})
		job.MarkCompleted()

		status := job.GetStatus()

		require.NotNil(t, job.CompletedAt)
		require.NotNil(t, status.CompletedAt)
		assert.Equal(t, *job.CompletedAt, *status.CompletedAt)

		// Verify they're separate pointers
		assert.NotSame(t, job.CompletedAt, status.CompletedAt, "CompletedAt should be copied, not shared")
	})

	t.Run("Empty results map is copied correctly", func(t *testing.T) {
		jq := NewJobQueue()
		job := jq.CreateJob([]string{"file1.mp4"})

		status := job.GetStatus()

		assert.Empty(t, status.Results)
		assert.NotNil(t, status.Results)
	})
}

// TestConcurrent_GetStatusAndUpdateFileResult validates thread-safe snapshot access
// This test catches race conditions where handlers read job state while workers update it
func TestConcurrent_GetStatusAndUpdateFileResult(t *testing.T) {
	jq := NewJobQueue()
	job := jq.CreateJob([]string{"file1.mp4", "file2.mkv", "file3.avi"})

	now := time.Now()
	// Initialize with a file result
	job.UpdateFileResult("file1.mp4", &FileResult{
		FilePath:  "file1.mp4",
		MovieID:   "IPX-100",
		Status:    JobStatusRunning,
		StartedAt: now,
	})

	// Simulate worker updating job results concurrently
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			// Rapidly update multiple file results
			job.UpdateFileResult("file1.mp4", &FileResult{
				FilePath:  "file1.mp4",
				MovieID:   "IPX-" + fmt.Sprintf("%d", i),
				Status:    JobStatusRunning,
				StartedAt: now,
			})
			job.UpdateFileResult("file2.mkv", &FileResult{
				FilePath:  "file2.mkv",
				MovieID:   "IPX-" + fmt.Sprintf("%d", i+1000),
				Status:    JobStatusCompleted,
				StartedAt: now,
			})
		}
		close(done)
	}()

	// Simulate handler reading job state concurrently (the safe pattern)
	for i := 0; i < 1000; i++ {
		// GetStatus() returns a thread-safe snapshot
		status := job.GetStatus()

		// Iterate over results (safe because it's a copy)
		for filePath, result := range status.Results {
			// Verify basic invariants
			assert.NotEmpty(t, filePath)
			if result != nil {
				assert.NotEmpty(t, result.FilePath)
			}
		}
	}

	<-done
}

// TestConcurrent_DirectMapAccessIsUnsafe demonstrates the race condition
// This test would fail with -race if we directly accessed job.Results without GetStatus()
// Run with: go test -race -run TestConcurrent_DirectMapAccessIsUnsafe
func TestConcurrent_DirectMapAccessIsUnsafe(t *testing.T) {
	t.Skip("This test demonstrates unsafe pattern - skip to avoid race detector failures")

	jq := NewJobQueue()
	job := jq.CreateJob([]string{"file1.mp4"})

	now := time.Now()
	job.UpdateFileResult("file1.mp4", &FileResult{
		FilePath:  "file1.mp4",
		MovieID:   "IPX-1",
		Status:    JobStatusRunning,
		StartedAt: now,
	})

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			job.UpdateFileResult("file1.mp4", &FileResult{
				FilePath:  "file1.mp4",
				MovieID:   fmt.Sprintf("IPX-%d", i),
				Status:    JobStatusRunning,
				StartedAt: now,
			})
		}
		close(done)
	}()

	// UNSAFE: Direct map access without GetStatus() - WOULD FAIL WITH -race
	for i := 0; i < 1000; i++ {
		// This would cause: fatal error: concurrent map iteration and map write
		for filePath := range job.Results {
			_ = filePath
		}
	}

	<-done
}

// TestBatchJob_PointerFieldIndependence validates that pointer fields are deep copied
// This ensures modifying pointer fields in the snapshot doesn't affect the live job
func TestBatchJob_PointerFieldIndependence(t *testing.T) {
	jq := NewJobQueue()
	job := jq.CreateJob([]string{"file1.mp4"})

	// Create FileResult with pointer fields
	originalTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	originalError := "poster download failed"
	job.UpdateFileResult("file1.mp4", &FileResult{
		FilePath:    "file1.mp4",
		MovieID:     "IPX-100",
		Status:      JobStatusCompleted,
		StartedAt:   time.Now(),
		EndedAt:     &originalTime,
		PosterError: &originalError,
	})

	// Get snapshot
	snapshot := job.GetStatus()

	// Verify initial values in snapshot
	assert.NotNil(t, snapshot.Results["file1.mp4"])
	assert.NotNil(t, snapshot.Results["file1.mp4"].EndedAt)
	assert.NotNil(t, snapshot.Results["file1.mp4"].PosterError)
	assert.Equal(t, originalTime, *snapshot.Results["file1.mp4"].EndedAt)
	assert.Equal(t, originalError, *snapshot.Results["file1.mp4"].PosterError)

	// Modify pointer fields in the snapshot
	newTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	newError := "different error"
	snapshot.Results["file1.mp4"].EndedAt = &newTime
	snapshot.Results["file1.mp4"].PosterError = &newError

	// Get fresh snapshot to verify original is unchanged
	freshSnapshot := job.GetStatus()
	assert.NotNil(t, freshSnapshot.Results["file1.mp4"])
	assert.NotNil(t, freshSnapshot.Results["file1.mp4"].EndedAt)
	assert.NotNil(t, freshSnapshot.Results["file1.mp4"].PosterError)

	// Verify original values are preserved (not affected by first snapshot modifications)
	assert.Equal(t, originalTime, *freshSnapshot.Results["file1.mp4"].EndedAt, "EndedAt should not be affected by snapshot modification")
	assert.Equal(t, originalError, *freshSnapshot.Results["file1.mp4"].PosterError, "PosterError should not be affected by snapshot modification")

	// Verify modified snapshot has new values
	assert.Equal(t, newTime, *snapshot.Results["file1.mp4"].EndedAt)
	assert.Equal(t, newError, *snapshot.Results["file1.mp4"].PosterError)
}
