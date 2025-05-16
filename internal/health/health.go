package health

import (
	"greenvue/internal/db"
	"greenvue/lib/errors"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
)

// System information
type SystemInfo struct {
	GoVersion  string    `json:"goVersion"`
	NumCPU     int       `json:"numCPU"`
	GoRoutines int       `json:"goRoutines"`
	MemStats   MemStats  `json:"memStats"`
	StartTime  time.Time `json:"startTime"`
	Uptime     string    `json:"uptime"`
}

// Memory statistics
type MemStats struct {
	Alloc      uint64 `json:"alloc"`      // Currently allocated memory in bytes
	TotalAlloc uint64 `json:"totalAlloc"` // Total allocated memory since start
	Sys        uint64 `json:"sys"`        // Total memory obtained from the OS
	NumGC      uint32 `json:"numGC"`      // Number of completed GC cycles
}

var startTime = time.Now()

// HealthCheck returns basic system health information
func HealthCheck(c *fiber.Ctx) error {
	logger := errors.GetRequestLogger(c)
	logger.Info("Health check requested")

	return errors.SuccessResponse(c, fiber.Map{
		"status": "UP",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func DetailedHealth(c *fiber.Ctx) error {
	// Get system info
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	info := SystemInfo{
		GoVersion:  runtime.Version(),
		NumCPU:     runtime.NumCPU(),
		GoRoutines: runtime.NumGoroutine(),
		MemStats: MemStats{
			Alloc:      stats.Alloc,
			TotalAlloc: stats.TotalAlloc,
			Sys:        stats.Sys,
			NumGC:      stats.NumGC,
		},
		StartTime: startTime,
		Uptime:    time.Since(startTime).String(),
	}

	// Check database connection and measure latency
	dbStatus := "UP"
	dbDetails := "Connected"
	var dbLatencyMs int64 = -1

	client := db.GetGlobalClient()

	start := time.Now()
	_, err := client.GET("listings", "select=*")
	dbLatencyMs = time.Since(start).Milliseconds()

	if err != nil {
		dbStatus = "DOWN"
		dbDetails = "Connection failed: " + err.Error()
	}

	return errors.SuccessResponse(c, fiber.Map{
		"status": "UP",
		"database": fiber.Map{
			"status":    dbStatus,
			"details":   dbDetails,
			"latencyMs": dbLatencyMs,
		},
		"system": info,
	})
}
