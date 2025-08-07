package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/travboz/greenlightv3/internal/data"
	"github.com/travboz/greenlightv3/internal/data/validator"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the payload struct to a new Movie struct.
	movie := &data.Movie{
		Title:   payload.Title,
		Year:    payload.Year,
		Runtime: payload.Runtime,
		Genres:  payload.Genres,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed. If they did, then use
	// the failedValidationResponse() helper to send a response to the client, passing
	// in the v.Errors map.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", payload)
}

// Add a showMovieHandler for the "GET /v1/movies/:id" endpoint. For now, we retrieve
// the interpolated "id" parameter from the current URL and include it in a placeholder
// response.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casino Royale",
		Runtime:   data.Runtime(144),
		Genres:    []string{"action", "thriller", "spy"},
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
