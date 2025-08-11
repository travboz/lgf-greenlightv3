package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// func (app *application) serve() error {
// 	srv := &http.Server{
// 		Addr:         fmt.Sprintf(":%d", app.config.port),
// 		Handler:      app.routes(),
// 		IdleTimeout:  time.Minute,
// 		ReadTimeout:  5 * time.Second,
// 		WriteTimeout: 10 * time.Second,
// 		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
// 	}

// 	// Create a shutdownError channel. We will use this to receive any errors returned
// 	// by the graceful Shutdown() function.
// 	shutdownError := make(chan error)

// 	// Start a background goroutine.
// 	go func() {
// 		// Create a quit channel which carries os.Signal values
// 		quit := make(chan os.Signal, 1)

// 		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and
// 		// relay them to the quit channel. Any other signals will not be caught by
// 		// signal.Notify() and will retain their default behavior.
// 		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// 		// Read the signal from the quit channel. This code will block until a signal is
// 		// received.
// 		s := <-quit
// 		// The code below WILL NOT RUN until we receive a SIGINT or SIGTERM signal.

// 		// Log a message to say that indicates the signal has been caught. Notice that we also
// 		// call the String() method on the signal to get the signal name and include it
// 		// in the log entry attributes.
// 		app.logger.Info("shutting down server", "signal", s.String())

// 		// Create a context with a 30-second timeout.
// 		// Gives our inflight requests 30 seconds to complete.
// 		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 		defer cancel()

// 		// Call Shutdown() on our server, passing in the context we just made.
// 		// Shutdown() will return nil if the graceful shutdown was successful, or an
// 		// error (which may happen because of a problem closing the listeners, or
// 		// because the shutdown didn't complete before the 30-second context deadline is
// 		// hit). We relay this return value to the shutdownError channel.
// 		// Call Shutdown() on the server like before, but now we only send on the
// 		// shutdownError channel if it returns an error.
// 		err := srv.Shutdown(ctx)
// 		if err != nil {
// 			shutdownError <- err
// 		}

// 		// Log a message to say that we're waiting for any background goroutines to
// 		// complete their tasks.
// 		app.logger.Info("completing background tasks", "addr", srv.Addr)

// 		// Call Wait() to block until our WaitGroup counter is zero --- essentially
// 		// blocking until the background goroutines have finished. Then we return nil on
// 		// the shutdownError channel, to indicate that the shutdown completed without
// 		// any issues.
// 		app.bwg.Wait()
// 		shutdownError <- nil
// 	}()

// 	// Log a "starting server" message
// 	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

// 	// Calling Shutdown() on our server will cause ListenAndServe() to immediately
// 	// return a http.ErrServerClosed error. So if we see this error, it is actually a
// 	// good thing and an indication that the graceful shutdown has started. So we check
// 	// specifically for this, only returning the error if it is NOT http.ErrServerClosed.
// 	err := srv.ListenAndServe()
// 	if !errors.Is(err, http.ErrServerClosed) {
// 		return err
// 	}

// 	// Otherwise, we wait to receive the return value from Shutdown() on the
// 	// shutdownError channel. If return value is an error, we know that there was a
// 	// problem with the graceful shutdown and we return the error.
// 	err = <-shutdownError
// 	if err != nil {
// 		return err
// 	}

// 	// At this point we know that the graceful shutdown completed successfully and we
// 	// log a "stopped server" message.
// 	app.logger.Info("stopped server", "addr", srv.Addr)

// 	return nil
// }

func (app *application) serve() error {
	// Create an HTTP server with address, routes, timeouts, and error logging
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),                      // Port to listen on
		Handler:      app.routes(),                                             // HTTP handler (routes)
		IdleTimeout:  time.Minute,                                              // Max keep-alive time for idle connections
		ReadTimeout:  5 * time.Second,                                          // Max time to read the request
		WriteTimeout: 10 * time.Second,                                         // Max time to write the response
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError), // Log only errors
	}

	// Channel for any errors that occur during graceful shutdown
	shutdownError := make(chan error)

	// Run a goroutine that waits for shutdown signals (SIGINT, SIGTERM)
	go func() {
		quit := make(chan os.Signal, 1) // Will store OS signals

		// Listen for interrupt (CTRL+C) or terminate signals
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Block until a signal is received
		s := <-quit

		// Log the caught signal
		app.logger.Info("shutting down server", "signal", s.String())

		// Create a 30-second timeout context for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt to gracefully stop the server
		// (Stop new requests, finish ongoing ones)
		if err := srv.Shutdown(ctx); err != nil {
			shutdownError <- err // Send error if shutdown failed
		}

		// Wait for any background goroutines to finish
		app.logger.Info("completing background tasks", "addr", srv.Addr)
		app.bwg.Wait()

		// Indicate shutdown finished with no issues
		shutdownError <- nil
	}()

	// Log that the server is starting
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// Start listening for requests
	// If Shutdown() is called, ListenAndServe() will return http.ErrServerClosed
	err := srv.ListenAndServe()

	// Only treat errors as real problems if they're not due to shutdown
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Wait for shutdown goroutine to finish and check for shutdown errors
	if err := <-shutdownError; err != nil {
		return err
	}

	// Log that the server has stopped successfully
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
