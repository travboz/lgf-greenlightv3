package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/travboz/greenlightv3/internal/data"
	"github.com/travboz/greenlightv3/internal/mailer"

	"github.com/julienschmidt/httprouter"
)

func newTestApp() *application {
	return &application{
		config: config{},
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		models: data.Models{},
		mailer: mailer.Mailer{},
	}

}

func TestReadIDParam(t *testing.T) {
	app := newTestApp()

	tests := []struct {
		name        string
		idParam     string
		expectedID  int64
		expectError bool
	}{
		{
			name:        "valid positive ID",
			idParam:     "123",
			expectedID:  123,
			expectError: false,
		},
		{
			name:        "valid large ID",
			idParam:     "9223372036854775807", // max int64
			expectedID:  9223372036854775807,
			expectError: false,
		},
		{
			name:        "zero ID should return error",
			idParam:     "0",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "negative ID should return error",
			idParam:     "-1",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "non-numeric string should return error",
			idParam:     "abc",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "empty string should return error",
			idParam:     "",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "float number should return error",
			idParam:     "123.45",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "number with spaces should return error",
			idParam:     " 123 ",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "very large number (overflow) should return error",
			idParam:     "99999999999999999999999999999999",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "hexadecimal number should return error",
			idParam:     "0x123",
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Create httprouter.Params and add to context
			params := httprouter.Params{
				httprouter.Param{Key: "id", Value: tt.idParam},
			}
			ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
			req = req.WithContext(ctx)

			// Call the function
			id, err := app.readIDParam(req)

			// Check results
			if tt.expectError {

				if err == nil {
					t.Errorf("wanted error but got none")
				}

				if id != 0 {
					t.Errorf("wanted id to be 0 when error occurs, got %d", id)
				}
			} else { // expect no error

				if err != nil {
					t.Errorf("didn't want an error: %v", err)
				}

				if id != tt.expectedID {
					t.Errorf("wanted id %d, got %d", tt.expectedID, id)
				}
			}
		})
	}
}

// func TestWriteJSON(t *testing.T) {
// 	app := newTestApp()

// 	t.Run("successful write of name", func(t *testing.T) {
// 		// Simulate a ResponseWriter.
// 		rr := httptest.NewRecorder()

// 		headers := http.Header{}
// 		headers.Add("Server", "Go")

// 		err := app.writeJSON(rr, http.StatusOK, envelope{"key": "value"}, headers)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		// Inspect the response.
// 		res := rr.Result()
// 		defer res.Body.Close()

// 		if got := res.StatusCode; got != http.StatusOK {
// 			t.Errorf("wanted status %d, got %d", http.StatusOK, got)
// 		}

// 		// Check Content-Type header
// 		if ct := res.Header.Get("Content-Type"); ct != "application/json" {
// 			t.Errorf("wanted Content-Type application/json, got %s", ct)
// 		}

// 		// Check envelope
// 		var body envelope
// 		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
// 			t.Fatalf("failed to decode response body: %v", err)
// 		}

// 		if got := body["key"]; got != "value" {
// 			t.Errorf("wanted value to be 'value', got %v", got)
// 		}

// 		// --- Check custom headers ---
// 		if server := res.Header.Get("Server"); server != "Go" {
// 			t.Errorf("wanted Server header 'Go', got %q", server)
// 		}

// 	})
// }

func TestWriteJSON(t *testing.T) {
	app := newTestApp()

	tests := []struct {
		name     string
		status   int
		data     envelope
		headers  http.Header
		wantErr  bool
		wantBody envelope
	}{
		{
			name:     "normal case with custom header",
			status:   http.StatusOK,
			data:     envelope{"key": "value"},
			headers:  http.Header{"Server": []string{"Go"}},
			wantErr:  false,
			wantBody: envelope{"key": "value"},
		},
		{
			name:     "empty headers",
			status:   http.StatusOK,
			data:     envelope{"foo": "bar"},
			headers:  nil,
			wantErr:  false,
			wantBody: envelope{"foo": "bar"},
		},
		{
			name:     "different status code",
			status:   http.StatusCreated,
			data:     envelope{"created": true},
			headers:  nil,
			wantErr:  false,
			wantBody: envelope{"created": true},
		},
		{
			name:    "marshal error",
			status:  http.StatusOK,
			data:    envelope{"bad": make(chan int)}, // cannot marshal
			headers: nil,
			wantErr: true,
		},
		{
			name:     "content-type override",
			status:   http.StatusOK,
			data:     envelope{"a": 1},
			headers:  http.Header{"Content-Type": []string{"text/plain"}},
			wantErr:  false,
			wantBody: envelope{"a": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			err := app.writeJSON(rr, tt.status, tt.data, tt.headers)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			res := rr.Result()
			defer res.Body.Close()

			if got := res.StatusCode; got != tt.status {
				t.Errorf("wanted status %d, got %d", tt.status, got)
			}

			if ct := res.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("wanted Content-Type application/json, got %s", ct)
			}

			if tt.wantBody != nil {
				var body envelope
				if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				for k, v := range tt.wantBody {
					got := body[k]
					// Compare numbers carefully: JSON numbers become float64
					switch wantVal := v.(type) {
					case int:
						if gotFloat, ok := got.(float64); !ok || int(gotFloat) != wantVal {
							t.Errorf("body[%q] = %v, want %v", k, got, wantVal)
						}
					default:
						if got != wantVal {
							t.Errorf("body[%q] = %v, want %v", k, got, wantVal)
						}
					}
				}
			}
		})
	}
}
