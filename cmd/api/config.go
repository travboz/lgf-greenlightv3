package main

import (
	"time"

	"github.com/joho/godotenv"
)

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the
// application (development, staging, production, etc.). We will read in these
// configuration settings from command-line flags when the application starts.
// Add db struct field which holds configuration settings for our db connection pool.
type config struct {
	port    int
	env     string
	db      dbConfig
	limiter limiterConfig
}

type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

// Add a new limiter struct containing fields for the requests-per-second and burst
// values, and a boolean field which we can use to enable/disable rate limiting
// altogether.
type limiterConfig struct {
	rps     float64 // refills per second
	burst   int
	enabled bool
}

func LoadEnv() error {
	return godotenv.Load()
}
