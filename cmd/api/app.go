package main

import (
	"log/slog"

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
}
