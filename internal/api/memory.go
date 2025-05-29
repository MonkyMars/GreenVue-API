package api

import (
	"runtime"
	"time"

	"greenvue/lib/errors"
	"greenvue/lib/image"

	"github.com/gofiber/fiber/v2"
)

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	// Go runtime memory stats
	Alloc      uint64 `json:"alloc_bytes"`       // Bytes currently allocated
	TotalAlloc uint64 `json:"total_alloc_bytes"` // Total bytes allocated over lifetime
	Sys        uint64 `json:"sys_bytes"`         // Bytes obtained from system
	NumGC      uint32 `json:"num_gc"`            // Number of garbage collections

	// Human readable versions
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`

	// Image queue stats
	PendingImages   int `json:"pending_images"`
	CompletedImages int `json:"completed_images"`

	// Goroutine count
	NumGoroutines int `json:"num_goroutines"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

// GetMemoryStats returns current memory usage statistics
func GetMemoryStats(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Force garbage collection to get accurate stats
	runtime.GC()

	stats := MemoryStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		NumGC:         m.NumGC,
		AllocMB:       float64(m.Alloc) / 1024 / 1024,
		TotalAllocMB:  float64(m.TotalAlloc) / 1024 / 1024,
		SysMB:         float64(m.Sys) / 1024 / 1024,
		NumGoroutines: runtime.NumGoroutine(),
		Timestamp:     time.Now(),
	}

	// Add image queue stats if available
	if image.GlobalImageQueue != nil {
		stats.PendingImages = image.GlobalImageQueue.PendingCount()
		stats.CompletedImages = image.GlobalImageQueue.GetCompletedJobsCount()
	}

	return errors.SuccessResponse(c, stats)
}

// ForceImageCleanup manually triggers cleanup of completed image jobs
func ForceImageCleanup(c *fiber.Ctx) error {
	if image.GlobalImageQueue == nil {
		return errors.BadRequest("Image queue is not initialized")
	}

	beforeCount := image.GlobalImageQueue.GetCompletedJobsCount()
	image.GlobalImageQueue.ForceCleanup()
	afterCount := image.GlobalImageQueue.GetCompletedJobsCount()

	return errors.SuccessResponse(c, fiber.Map{
		"message":         "Cleanup completed",
		"jobs_before":     beforeCount,
		"jobs_after":      afterCount,
		"jobs_cleaned_up": beforeCount - afterCount,
	})
}
