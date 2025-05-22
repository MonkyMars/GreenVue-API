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
	jobs     map[string]*Job
	mu       sync.RWMutex
	shutdown chan struct{}
}

// NewScheduler creates a new job scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:     make(map[string]*Job),
		shutdown: make(chan struct{}),
	}
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(id, name, description string, fn JobFunc, interval time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

// runJob executes a job on its schedule
func (s *Scheduler) runJob(job *Job) {
	for {
		select {
		case <-job.ctx.Done():
			return // Exit if job is cancelled
		case <-time.After(time.Until(job.NextRun)):
			func() {
				s.mu.Lock()
				job.IsRunning = true
				job.LastRun = time.Now()
				s.mu.Unlock() // Execute the job
				err := job.Func(job.ctx)
				if err != nil {
					log.Printf("Error executing job %s: %v", job.ID, err)
				}

				s.mu.Lock()
				job.IsRunning = false
				job.NextRun = time.Now().Add(job.Interval)
				s.mu.Unlock()
			}()
		}
	}
}

// Shutdown gracefully shuts down the scheduler
func (s *Scheduler) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel all jobs
	for _, job := range s.jobs {
		job.cancel()
	}

	close(s.shutdown)
}
