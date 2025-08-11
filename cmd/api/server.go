package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Log a "starting server" message
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// Start server as normal, returning the error
	return srv.ListenAndServe()
}
