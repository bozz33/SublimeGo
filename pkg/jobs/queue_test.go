package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue(4)
	require.NotNil(t, q)
	assert.Equal(t, 4, q.workers)
	assert.NotNil(t, q.jobChan)
}

func TestQueueStartStop(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	assert.True(t, q.started)

	q.Stop()
	assert.True(t, q.started) // Stays true after stop
}

func TestDispatch(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	executed := false
	jobID := q.Dispatch("test-job", func(ctx context.Context, job *Job) error {
		executed = true
		return nil
	})

	assert.NotEmpty(t, jobID)

	// Wait for job execution
	time.Sleep(100 * time.Millisecond)
	assert.True(t, executed)

	job, ok := q.Get(jobID)
	require.True(t, ok)
	assert.Equal(t, StatusCompleted, job.Status)
}

func TestDispatchWithError(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	expectedErr := errors.New("test error")
	jobID := q.Dispatch("failing-job", func(ctx context.Context, job *Job) error {
		return expectedErr
	})

	// Wait for job execution
	time.Sleep(100 * time.Millisecond)

	job, ok := q.Get(jobID)
	require.True(t, ok)
	assert.Equal(t, StatusFailed, job.Status)
	assert.Equal(t, expectedErr, job.Error)
}

func TestDispatchWithCallbacks(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	completeCalled := false
	errorCalled := false

	jobID := q.DispatchWithCallbacks(
		"callback-job",
		func(ctx context.Context, job *Job) error {
			return nil
		},
		func(job *Job) {
			completeCalled = true
		},
		func(job *Job, err error) {
			errorCalled = true
		},
	)

	time.Sleep(100 * time.Millisecond)

	assert.True(t, completeCalled)
	assert.False(t, errorCalled)

	job, ok := q.Get(jobID)
	require.True(t, ok)
	assert.Equal(t, StatusCompleted, job.Status)
}

func TestDispatchWithErrorCallback(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	completeCalled := false
	errorCalled := false
	var capturedErr error

	q.DispatchWithCallbacks(
		"error-callback-job",
		func(ctx context.Context, job *Job) error {
			return errors.New("test error")
		},
		func(job *Job) {
			completeCalled = true
		},
		func(job *Job, err error) {
			errorCalled = true
			capturedErr = err
		},
	)

	time.Sleep(100 * time.Millisecond)

	assert.False(t, completeCalled)
	assert.True(t, errorCalled)
	assert.NotNil(t, capturedErr)
}

func TestGet(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	jobID := q.Dispatch("test-job", func(ctx context.Context, job *Job) error {
		return nil
	})

	job, ok := q.Get(jobID)
	require.True(t, ok)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, "test-job", job.Name)
}

func TestGetAll(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	q.Dispatch("job1", func(ctx context.Context, job *Job) error { return nil })
	q.Dispatch("job2", func(ctx context.Context, job *Job) error { return nil })
	q.Dispatch("job3", func(ctx context.Context, job *Job) error { return nil })

	time.Sleep(100 * time.Millisecond)

	jobs := q.GetAll()
	assert.Len(t, jobs, 3)
}

func TestGetByStatus(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	q.Dispatch("success-job", func(ctx context.Context, job *Job) error { return nil })
	q.Dispatch("fail-job", func(ctx context.Context, job *Job) error { return errors.New("error") })

	time.Sleep(100 * time.Millisecond)

	completed := q.GetByStatus(StatusCompleted)
	failed := q.GetByStatus(StatusFailed)

	assert.Len(t, completed, 1)
	assert.Len(t, failed, 1)
}

func TestCount(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	assert.Equal(t, 0, q.Count())

	q.Dispatch("job1", func(ctx context.Context, job *Job) error { return nil })
	q.Dispatch("job2", func(ctx context.Context, job *Job) error { return nil })

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 2, q.Count())
}

func TestCountByStatus(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	q.Dispatch("job1", func(ctx context.Context, job *Job) error { return nil })
	q.Dispatch("job2", func(ctx context.Context, job *Job) error { return errors.New("error") })

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, q.CountByStatus(StatusCompleted))
	assert.Equal(t, 1, q.CountByStatus(StatusFailed))
}

func TestWait(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	jobID := q.Dispatch("slow-job", func(ctx context.Context, job *Job) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	job, err := q.Wait(jobID, 1*time.Second)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, job.Status)
}

func TestWaitTimeout(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	jobID := q.Dispatch("very-slow-job", func(ctx context.Context, job *Job) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	_, err := q.Wait(jobID, 100*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestCancel(t *testing.T) {
	q := NewQueue(1) // 1 worker to control execution
	q.Start()
	defer q.Stop()

	// Block the worker with a slow job
	q.Dispatch("blocking-job", func(ctx context.Context, job *Job) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	// Add a pending job
	jobID := q.Dispatch("pending-job", func(ctx context.Context, job *Job) error {
		return nil
	})

	time.Sleep(50 * time.Millisecond)

	// Job should be pending
	job, _ := q.Get(jobID)
	assert.Equal(t, StatusPending, job.Status)

	// Cancel the job
	err := q.Cancel(jobID)
	assert.NoError(t, err)

	job, _ = q.Get(jobID)
	assert.Equal(t, StatusCancelled, job.Status)
}

func TestClear(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	// Create completed jobs
	q.Dispatch("job1", func(ctx context.Context, job *Job) error { return nil })
	q.Dispatch("job2", func(ctx context.Context, job *Job) error { return nil })

	time.Sleep(100 * time.Millisecond)

	// Manually modify CompletedAt to simulate old jobs
	jobs := q.GetAll()
	for _, job := range jobs {
		old := time.Now().Add(-2 * time.Hour)
		job.CompletedAt = &old
		q.jobs.Store(job.ID, job)
	}

	deleted := q.Clear(1 * time.Hour)
	assert.Equal(t, 2, deleted)
	assert.Equal(t, 0, q.Count())
}

func TestStats(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	defer q.Stop()

	q.Dispatch("job1", func(ctx context.Context, job *Job) error { return nil })
	q.Dispatch("job2", func(ctx context.Context, job *Job) error { return errors.New("error") })

	time.Sleep(100 * time.Millisecond)

	stats := q.Stats()
	assert.Equal(t, 2, stats["total"])
	assert.Equal(t, 1, stats["completed"])
	assert.Equal(t, 1, stats["failed"])
	assert.Equal(t, 2, stats["workers"])
}

func TestJobUpdateProgress(t *testing.T) {
	job := &Job{}

	job.UpdateProgress(50)
	assert.Equal(t, 50, job.Progress)

	job.UpdateProgress(150) // Should be capped at 100
	assert.Equal(t, 100, job.Progress)

	job.UpdateProgress(-10) // Should be capped at 0
	assert.Equal(t, 0, job.Progress)
}

func TestJobSetResult(t *testing.T) {
	job := &Job{}
	result := map[string]string{"status": "success"}

	job.SetResult(result)
	assert.Equal(t, result, job.Result)
}

func TestJobDuration(t *testing.T) {
	job := &Job{}

	// Not started yet
	assert.Equal(t, time.Duration(0), job.Duration())

	// Started but not completed
	now := time.Now()
	job.StartedAt = &now
	time.Sleep(10 * time.Millisecond)
	assert.Greater(t, job.Duration(), time.Duration(0))

	// Completed
	completed := time.Now()
	job.CompletedAt = &completed
	duration := job.Duration()
	assert.Greater(t, duration, time.Duration(0))
}

func TestJobIsCompleted(t *testing.T) {
	job := &Job{}

	job.Status = StatusPending
	assert.False(t, job.IsCompleted())

	job.Status = StatusRunning
	assert.False(t, job.IsCompleted())

	job.Status = StatusCompleted
	assert.True(t, job.IsCompleted())

	job.Status = StatusFailed
	assert.True(t, job.IsCompleted())

	job.Status = StatusCancelled
	assert.True(t, job.IsCompleted())
}

func TestJobIsSuccess(t *testing.T) {
	job := &Job{}

	job.Status = StatusCompleted
	assert.True(t, job.IsSuccess())

	job.Status = StatusFailed
	assert.False(t, job.IsSuccess())
}

func TestJobIsFailed(t *testing.T) {
	job := &Job{}

	job.Status = StatusFailed
	assert.True(t, job.IsFailed())

	job.Status = StatusCompleted
	assert.False(t, job.IsFailed())
}

func BenchmarkDispatch(b *testing.B) {
	q := NewQueue(4)
	q.Start()
	defer q.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Dispatch("bench-job", func(ctx context.Context, job *Job) error {
			return nil
		})
	}
}
