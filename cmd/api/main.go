package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/travboz/greenlightv3/internal/data"
	"github.com/travboz/greenlightv3/internal/mailer"
	"github.com/travboz/greenlightv3/internal/vcs"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
var (
	version = vcs.Version()
)

func main() {

	if err := LoadEnv(); err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := NewConfig()

	// If the version flag value is true, then print out the version number and
	// immediately exit.
	if cfg.displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	// Initialize a new structured logger which writes log entries to the standard out
	// stream.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	publishMetrics(db)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// Start server with app.serve()
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
