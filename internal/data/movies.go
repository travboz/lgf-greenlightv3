package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`                // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                 // Timestamp for when the movie is added to our database; hide it so it doesn't show in the json output
	Title     string    `json:"title"`             // Movie title
	Year      int32     `json:"year,omitempty"`    // Movie release year; omit if empty
	Runtime   Runtime   `json:"runtime,omitempty"` // Movie runtime (in minutes); omit if empty
	Genres    []string  `json:"genres,omitempty"`  // Slice of genres for the movie (romance, comedy, etc.); omit if empty
	Version   int32     `json:"version"`           // The version number starts at 1 and will be incremented each time the movie information is updated
}
