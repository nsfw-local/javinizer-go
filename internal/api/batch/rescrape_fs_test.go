package batch

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCaseInsensitiveFS(t *testing.T) {
	tmpDir := t.TempDir()

	result := isCaseInsensitiveFS(tmpDir)

	t.Logf("Filesystem case-insensitive: %v", result)
}

func TestIsCaseInsensitiveFSCached_CachesResult(t *testing.T) {
	clearFSCaseCache()

	tmpDir := t.TempDir()

	result1 := isCaseInsensitiveFSCached(filepath.Join(tmpDir, "test1.jpg"))
	result2 := isCaseInsensitiveFSCached(filepath.Join(tmpDir, "test2.jpg"))

	assert.Equal(t, result1, result2, "cached results should match for same directory")
}

func TestIsCaseInsensitiveFS_HandlesNonexistentDir(t *testing.T) {
	clearFSCaseCache()

	result := isCaseInsensitiveFS("/nonexistent/path/that/does/not/exist")

	assert.False(t, result, "nonexistent directory should default to case-sensitive (safe)")
}

func clearFSCaseCache() {
	fsCaseCacheMu.Lock()
	fsCaseCache = make(map[string]bool)
	fsCaseCacheMu.Unlock()
}

type mockJobRepo struct {
	deleteErr error
}

func (m *mockJobRepo) Create(job *models.Job) error {
	return nil
}

func (m *mockJobRepo) Update(job *models.Job) error {
	return nil
}

func (m *mockJobRepo) Delete(id string) error {
	return m.deleteErr
}

func (m *mockJobRepo) FindByID(id string) (*models.Job, error) {
	return nil, nil
}

func (m *mockJobRepo) List() ([]models.Job, error) {
	return nil, nil
}

func (m *mockJobRepo) DeleteOrganizedOlderThan(date time.Time) error {
	return nil
}

var _ database.JobRepositoryInterface = (*mockJobRepo)(nil)

func TestDeleteJob_ReturnsErrorWhenDatabaseFails(t *testing.T) {
	mockRepo := &mockJobRepo{deleteErr: errors.New("database connection lost")}

	jq := worker.NewJobQueue(mockRepo, "")

	job := jq.CreateJob([]string{"/path/to/test.mp4"})
	job.MarkCompleted()

	tmpDir := t.TempDir()

	err := jq.DeleteJob(job.ID, tmpDir)

	require.Error(t, err, "expected error when database deletion fails")
	assert.True(t, strings.Contains(err.Error(), "database deletion failed"),
		"error should mention database failure, got: %v", err)
}

func TestDeleteJob_ReturnsErrorWhenJobNotFound(t *testing.T) {
	jq := worker.NewJobQueue(nil, "")

	tmpDir := t.TempDir()

	err := jq.DeleteJob("nonexistent-job", tmpDir)

	require.Error(t, err, "expected error when job not found")
	assert.True(t, strings.Contains(err.Error(), "not found"),
		"error should mention job not found, got: %v", err)
}

func TestDeleteJob_ReturnsErrorWhenJobRunning(t *testing.T) {
	jq := worker.NewJobQueue(nil, "")

	job := jq.CreateJob([]string{"/path/to/test.mp4"})
	job.MarkStarted()

	tmpDir := t.TempDir()

	err := jq.DeleteJob(job.ID, tmpDir)

	require.Error(t, err, "expected error when deleting running job")
	assert.True(t, strings.Contains(err.Error(), "running"),
		"error should mention running job, got: %v", err)
}
