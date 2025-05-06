package db

import (
	"os"
	"fmt"
)

func SanityCheck() (bool, error) {

	client := NewSupabaseClient() 

	if client == nil {
		return false, fmt.Errorf("failed to create Supabase client")
	}

	envs := [5]string {
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
};