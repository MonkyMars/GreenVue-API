package errors

import (
	stderrors "errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter provides IP-based rate limiting with customizable settings
type RateLimiter struct {
	// Maximum number of requests allowed in a time window
	Max int
	// Duration of the rate limiting window
	Expiration time.Duration
	// Function to extract the client identifier (default is IP)
	KeyGenerator func(*fiber.Ctx) string
	// Option to skip certain requests (e.g., for whitelisted IPs)
	SkipFunc func(*fiber.Ctx) bool

	storage    map[string]rateLimitEntry
	mutex      sync.RWMutex
	cleanupDue time.Time
}

type rateLimitEntry struct {
	count    int
	expireAt time.Time
}

// NewRateLimiter creates a new rate limiter with default settings
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		Max:        100,              // Default: 100 requests
		Expiration: 60 * time.Second, // Default: 60 second window
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Default: use client IP
		},
		SkipFunc: func(c *fiber.Ctx) bool {
			return false // Default: don't skip any requests
		},
		storage:    make(map[string]rateLimitEntry),
		cleanupDue: time.Now().Add(5 * time.Minute),
	}
}

// Middleware creates the Fiber middleware handler
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if necessary
		if rl.SkipFunc(c) {
			return c.Next()
		}

		// Get client identifier
		key := rl.KeyGenerator(c)

		// Check and update rate limit
		remaining, resetAt, limited := rl.check(key)

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(rl.Max))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if limited {
			// Get X-Request-ID for tracking
			requestID, _ := c.Locals("requestID").(string)
			if requestID == "" {
				requestID = "unknown"
			}

			// Log rate limit exceeded
			log.Printf("[%s] Rate limit exceeded for %s", requestID, key)

			// Return standard rate limit error
			return TooManyRequests("Rate limit exceeded. Please try again later.")
		}

		// Proceed with the request
		return c.Next()
	}
}

// check verifies and updates rate limit for a key
func (rl *RateLimiter) check(key string) (remaining int, resetAt time.Time, limited bool) {
	now := time.Now()

	// Perform cleanup if needed
	if now.After(rl.cleanupDue) {
		go rl.cleanup(now)
	}

	// Check and update rate limit
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	entry, exists := rl.storage[key]

	// If entry exists and is still valid
	if exists && now.Before(entry.expireAt) {
		entry.count++
		rl.storage[key] = entry

		// Check if rate limit exceeded
		if entry.count > rl.Max {
			return 0, entry.expireAt, true
		}

		return rl.Max - entry.count, entry.expireAt, false
	}

	// Create new entry
	expireAt := now.Add(rl.Expiration)
	rl.storage[key] = rateLimitEntry{
		count:    1,
		expireAt: expireAt,
	}

	return rl.Max - 1, expireAt, false
}

// cleanup removes expired entries
func (rl *RateLimiter) cleanup(now time.Time) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Remove expired entries
	for key, entry := range rl.storage {
		if now.After(entry.expireAt) {
			delete(rl.storage, key)
		}
	}

	// Schedule next cleanup
	rl.cleanupDue = now.Add(5 * time.Minute)
}

// TooManyRequests creates a 429 Too Many Requests error
func TooManyRequests(message string) *AppError {
	return New(stderrors.New("too many requests"), fiber.StatusTooManyRequests, message)
}
