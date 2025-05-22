package api

import (
	"greenvue/internal/jobs"
)

// GetJobScheduler returns the global job scheduler
func GetJobScheduler() *jobs.Scheduler {
	return jobs.GlobalScheduler
}
