package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

// Status represents the state of a job.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Job represents a task to execute.
type Job struct {
	ID          string
	Name        string
	Status      Status
	Progress    int // 0-100
	Result      interface{}
	Error       error
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Handler     func(ctx context.Context, job *Job) error
	OnComplete  func(job *Job)
	OnError     func(job *Job, err error)
}

// Queue manages asynchronous job execution.
type Queue struct {
	jobs    sync.Map // map[string]*Job
	workers int
	jobChan chan *Job
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
	started bool
	store   *Store // optional SQLite persistence
}

// NewQueue creates a new queue with a number of workers.
func NewQueue(workers int) *Queue {
	if workers <= 0 {
		workers = 4
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		workers: workers,
		jobChan: make(chan *Job, workers*10),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// NewPersistentQueue creates a queue backed by a SQLite store.
// Pending jobs from previous runs are automatically re-queued on Start().
func NewPersistentQueue(workers int, storePath string) (*Queue, error) {
	q := NewQueue(workers)

	store, err := NewStore(storePath)
	if err != nil {
		return nil, fmt.Errorf("jobs: create persistent queue: %w", err)
	}

	q.store = store
	return q, nil
}

// Start starts the queue workers.
// If the queue has a Store, pending jobs from previous runs are re-queued.
func (q *Queue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.started {
		return
	}

	q.started = true

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	if q.store != nil {
		pending, err := q.store.LoadPending()
		if err == nil {
			for _, job := range pending {
				q.jobs.Store(job.ID, job)
				if job.Handler != nil {
					q.jobChan <- job
				}
			}
		}
	}
}

// Stop stops the queue and waits for all jobs to finish.
func (q *Queue) Stop() {
	q.mu.Lock()
	if !q.started {
		q.mu.Unlock()
		return
	}
	q.mu.Unlock()

	close(q.jobChan)
	q.wg.Wait()
	q.cancel()
}

// worker processes jobs from the queue.
func (q *Queue) worker(id int) {
	defer q.wg.Done()

	for job := range q.jobChan {
		q.executeJob(job)
	}
}

// executeJob executes a job.
func (q *Queue) executeJob(job *Job) {
	job.Status = StatusRunning
	now := time.Now()
	job.StartedAt = &now
	q.jobs.Store(job.ID, job)
	q.persist(job)

	ctx, cancel := context.WithTimeout(q.ctx, 30*time.Minute)
	defer cancel()

	err := job.Handler(ctx, job)
	completed := time.Now()
	job.CompletedAt = &completed

	if err != nil {
		job.Status = StatusFailed
		job.Error = err
		if job.OnError != nil {
			job.OnError(job, err)
		}
	} else {
		job.Status = StatusCompleted
		job.Progress = 100
		if job.OnComplete != nil {
			job.OnComplete(job)
		}
	}

	q.jobs.Store(job.ID, job)
	q.persist(job)
}

// persist saves the job to the store if one is configured.
func (q *Queue) persist(job *Job) {
	if q.store != nil {
		_ = q.store.Save(job)
	}
}

// Dispatch adds a job to the queue.
func (q *Queue) Dispatch(name string, handler func(ctx context.Context, job *Job) error) string {
	job := &Job{
		ID:        uuid.New().String(),
		Name:      name,
		Status:    StatusPending,
		Progress:  0,
		CreatedAt: time.Now(),
		Handler:   handler,
	}

	q.jobs.Store(job.ID, job)
	q.persist(job)
	q.jobChan <- job

	return job.ID
}

// DispatchWithCallbacks adds a job with callbacks.
func (q *Queue) DispatchWithCallbacks(
	name string,
	handler func(ctx context.Context, job *Job) error,
	onComplete func(job *Job),
	onError func(job *Job, err error),
) string {
	job := &Job{
		ID:         uuid.New().String(),
		Name:       name,
		Status:     StatusPending,
		Progress:   0,
		CreatedAt:  time.Now(),
		Handler:    handler,
		OnComplete: onComplete,
		OnError:    onError,
	}

	q.jobs.Store(job.ID, job)
	q.persist(job)
	q.jobChan <- job

	return job.ID
}

// Get retrieves a job by its ID.
func (q *Queue) Get(id string) (*Job, bool) {
	value, ok := q.jobs.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*Job), true
}

// GetAll returns all jobs.
func (q *Queue) GetAll() []*Job {
	jobs := make([]*Job, 0)
	q.jobs.Range(func(key, value interface{}) bool {
		jobs = append(jobs, value.(*Job))
		return true
	})
	return jobs
}

// GetByStatus returns jobs with a given status.
func (q *Queue) GetByStatus(status Status) []*Job {
	return lo.Filter(q.GetAll(), func(job *Job, _ int) bool {
		return job.Status == status
	})
}

// Count returns the total number of jobs.
func (q *Queue) Count() int {
	count := 0
	q.jobs.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// CountByStatus returns the number of jobs by status.
func (q *Queue) CountByStatus(status Status) int {
	return len(q.GetByStatus(status))
}

// Wait waits for a job to finish (completed or failed).
func (q *Queue) Wait(id string, timeout time.Duration) (*Job, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for job %s", id)
		}

		job, ok := q.Get(id)
		if !ok {
			return nil, fmt.Errorf("job %s not found", id)
		}

		if job.Status == StatusCompleted || job.Status == StatusFailed || job.Status == StatusCancelled {
			return job, nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// Cancel cancels a pending job.
func (q *Queue) Cancel(id string) error {
	job, ok := q.Get(id)
	if !ok {
		return fmt.Errorf("job %s not found", id)
	}

	if job.Status != StatusPending {
		return fmt.Errorf("cannot cancel job %s with status %s", id, job.Status)
	}

	job.Status = StatusCancelled
	q.jobs.Store(id, job)

	return nil
}

// Clear removes completed jobs older than X time.
func (q *Queue) Clear(olderThan time.Duration) int {
	threshold := time.Now().Add(-olderThan)
	deleted := 0

	q.jobs.Range(func(key, value interface{}) bool {
		job := value.(*Job)
		if job.CompletedAt != nil && job.CompletedAt.Before(threshold) {
			q.jobs.Delete(key)
			deleted++
		}
		return true
	})

	return deleted
}

// Stats returns statistics about the queue.
func (q *Queue) Stats() map[string]interface{} {
	return map[string]interface{}{
		"total":     q.Count(),
		"pending":   q.CountByStatus(StatusPending),
		"running":   q.CountByStatus(StatusRunning),
		"completed": q.CountByStatus(StatusCompleted),
		"failed":    q.CountByStatus(StatusFailed),
		"cancelled": q.CountByStatus(StatusCancelled),
		"workers":   q.workers,
	}
}

// UpdateProgress updates the progress of a job.
func (j *Job) UpdateProgress(progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	j.Progress = progress
}

// SetResult sets the result of a job.
func (j *Job) SetResult(result interface{}) {
	j.Result = result
}

// Duration returns the execution duration of the job.
func (j *Job) Duration() time.Duration {
	if j.StartedAt == nil {
		return 0
	}
	if j.CompletedAt == nil {
		return time.Since(*j.StartedAt)
	}
	return j.CompletedAt.Sub(*j.StartedAt)
}

// IsCompleted checks if the job is finished (success or failure).
func (j *Job) IsCompleted() bool {
	return j.Status == StatusCompleted || j.Status == StatusFailed || j.Status == StatusCancelled
}

// IsSuccess checks if the job completed successfully.
func (j *Job) IsSuccess() bool {
	return j.Status == StatusCompleted
}

// IsFailed checks if the job failed.
func (j *Job) IsFailed() bool {
	return j.Status == StatusFailed
}
