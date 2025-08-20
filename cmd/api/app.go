package main

import (
	"database/sql"
	"log/slog"
	"sync"

	"github.com/travboz/greenlightv3/internal/data"
	"github.com/travboz/greenlightv3/internal/mailer"
)

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and a
// logger, but it will grow to include a lot more as our build progresses.
type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	bwg    sync.WaitGroup
	// Include a sync.WaitGroup in the application struct. The zero-value for a
	// sync.WaitGroup type is a valid, useable, sync.WaitGroup with a 'counter' value of 0,
	// so we don't need to do anything else to initialize it before we can use it.
}

func NewApplication(cfg config, logger *slog.Logger, db *sql.DB) *application {
	return &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
}
