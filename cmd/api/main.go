package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/travboz/greenlightv3/internal/vcs"
	"github.com/travboz/greenlightv3/pkg/env"
)

// Declare a string containing the application version number.
var (
	version = vcs.Version()
)

func main() {

	// Initialize a new structured logger which writes log entries to the standard out
	// stream.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if stage := env.GetString("ENV_STAGE", "development"); stage == "development" {
		if err := env.LoadEnv(); err != nil {
			logger.Error("Error loading .env file", "err", err.Error())
			os.Exit(1)
		}
	}

	cfg := NewConfig()

	// If the version flag value is true, then print out the version number and
	// immediately exit.
	if cfg.displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	publishMetrics(db)

	app := NewApplication(cfg, logger, db)

	// Start server with app.serve()
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
