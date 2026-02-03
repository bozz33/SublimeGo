// Package jobs provides an asynchronous job queue with worker pool.
//
// It allows dispatching background tasks that are processed by a pool of
// worker goroutines. Jobs support progress tracking, callbacks, cancellation,
// and timeout handling.
//
// Features:
//   - Worker pool with configurable size
//   - Job status tracking (pending, running, completed, failed)
//   - Progress reporting (0-100%)
//   - OnComplete and OnError callbacks
//   - Job cancellation
//   - Timeout handling (default 30 minutes)
//   - Job cleanup for old completed jobs
//
// Basic usage:
//
//	queue := jobs.NewQueue(4) // 4 workers
//	queue.Start()
//	defer queue.Stop()
//
//	// Dispatch a job
//	jobID := queue.Dispatch("send-email", func(ctx context.Context, job *jobs.Job) error {
//		job.UpdateProgress(50)
//		// Do work...
//		job.UpdateProgress(100)
//		return nil
//	})
//
//	// Wait for completion
//	result, err := queue.Wait(jobID, 5*time.Minute)
package jobs
