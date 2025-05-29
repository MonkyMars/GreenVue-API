package db

import (
	"fmt"
	"os"
)

func SanityCheck() (bool, error) {
	// Initialize the global client
	client, err := InitGlobalClient()
	if err != nil {
		return false, fmt.Errorf("failed to initialize global Supabase client: %w", err)
	}

	if client == nil {
		return false, fmt.Errorf("failed to create Supabase client")
	}

	query := fmt.Sprintf("select=*&limit=%s&order=created_at.desc", "1")
	_, err = client.GET("listing_details", query)

	if err != nil {
		return false, fmt.Errorf("failed to fetch listings: %w", err)
	}

	envs := [5]string{
		"SUPABASE_URL",
		"SUPABASE_ANON",
		"SUPABASE_SERVICE_KEY",
		"JWT_REFRESH_SECRET",
		"JWT_ACCESS_SECRET",
	}

	for _, env := range envs {
		if os.Getenv(env) == "" {
			return false, fmt.Errorf("environment variable %s is not set", env)
		}
	}

	return true, nil
}
