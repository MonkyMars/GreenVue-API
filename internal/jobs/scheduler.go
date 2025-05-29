package jobs

import (
	"context"
	"greenvue/lib/errors"
	"log"
	"sync"
	"time"
)

// JobFunc defines a function to be executed as a job
type JobFunc func(ctx context.Context) error

// Job represents a background job
type Job struct {
	ID          string
	Name        string
	Description string
	Func        JobFunc
	Interval    time.Duration
	LastRun     time.Time
	NextRun     time.Time
	IsRunning   bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// Scheduler manages background jobs
type Scheduler struct {
	jobs          map[string]*Job
	mu            sync.RWMutex
	shutdown      chan struct{}
	isShutdown    bool
	maxConcurrent int           // Limit concurrent job executions
	semaphore     chan struct{} // Semaphore for controlling concurrency
}

// NewScheduler creates a new job scheduler
func NewScheduler() *Scheduler {
	maxConcurrent := 5 // Limit to 5 concurrent jobs to prevent resource exhaustion
	return &Scheduler{
		jobs:          make(map[string]*Job),
		shutdown:      make(chan struct{}),
		isShutdown:    false,
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(id, name, description string, fn JobFunc, interval time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isShutdown {
		return errors.BadRequest("Scheduler is shutdown")
	}

	if _, exists := s.jobs[id]; exists {
		return errors.BadRequest("Job with this ID already exists")
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := &Job{
		ID:          id,
		Name:        name,
		Description: description,
		Func:        fn,
		Interval:    interval,
		NextRun:     time.Now().Add(interval),
		ctx:         ctx,
		cancel:      cancel,
	}

	s.jobs[id] = job
	go s.runJob(job)
	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return errors.BadRequest("Job not found")
	}

	job.cancel() // Cancel the context to stop the job
	delete(s.jobs, id)
	return nil
}

// GetJobs returns all jobs
func (s *Scheduler) GetJobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetJob returns a specific job by ID
func (s *Scheduler) GetJob(id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, errors.BadRequest("Job not found")
	}
	return job, nil
}

// runJob executes a job on its schedule with concurrency control
func (s *Scheduler) runJob(job *Job) {
	defer func() {
		// Ensure we clean up if there's a panic
		if r := recover(); r != nil {
			log.Printf("Job %s panicked: %v", job.ID, r)
		}
	}()

	for {
		select {
		case <-job.ctx.Done():
			return // Exit if job is cancelled
		case <-s.shutdown:
			return // Exit if scheduler is shutting down
		case <-time.After(time.Until(job.NextRun)):
			// Acquire semaphore before executing job
			select {
			case s.semaphore <- struct{}{}:
				// Got semaphore, execute job
				func() {
					defer func() {
						// Always release semaphore
						<-s.semaphore

						// Handle panics in job execution
						if r := recover(); r != nil {
							log.Printf("Job %s execution panicked: %v", job.ID, r)
						}
					}()

					s.mu.Lock()
					if s.isShutdown {
						s.mu.Unlock()
						return
					}
					job.IsRunning = true
					job.LastRun = time.Now()
					s.mu.Unlock()

					// Execute the job with timeout context
					jobCtx, cancel := context.WithTimeout(job.ctx, 30*time.Minute) // 30 minute timeout
					defer cancel()

					err := job.Func(jobCtx)
					if err != nil {
						log.Printf("Error executing job %s: %v", job.ID, err)
					}

					s.mu.Lock()
					if !s.isShutdown {
						job.IsRunning = false
						job.NextRun = time.Now().Add(job.Interval)
					}
					s.mu.Unlock()
				}()
			case <-job.ctx.Done():
				return // Exit if job is cancelled while waiting for semaphore
			case <-s.shutdown:
				return // Exit if scheduler is shutting down while waiting for semaphore
			}
		}
	}
}

// Shutdown gracefully shuts down the scheduler
func (s *Scheduler) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isShutdown {
		return // Already shutdown
	}

	s.isShutdown = true

	// Cancel all jobs
	for _, job := range s.jobs {
		job.cancel()
	}

	// Signal shutdown to all goroutines
	close(s.shutdown)

	// Give jobs a moment to finish gracefully
	time.Sleep(1 * time.Second)

	log.Printf("Scheduler shutdown complete, %d jobs stopped", len(s.jobs))
}
