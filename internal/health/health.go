package health

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"
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

// DetailedHealth returns detailed system health information
func DetailedHealth(c *fiber.Ctx) error {
	// Only allow this for authorized users
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return errors.Unauthorized("Authorization required for detailed health check")
	}

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

	// Check database connection
	dbStatus := "UP"
	dbDetails := "Connected"

	client := db.NewSupabaseClient()
	_, err := client.Query("listings", "select=count(*)")
	if err != nil {
		dbStatus = "DOWN"
		dbDetails = "Connection failed: " + err.Error()
	}

	return errors.SuccessResponse(c, fiber.Map{
		"status": "UP",
		"database": fiber.Map{
			"status":  dbStatus,
			"details": dbDetails,
		},
		"system": info,
	})
}
