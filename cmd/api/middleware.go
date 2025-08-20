package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pascaldekloe/jwt"
	"github.com/tomasen/realip"
	"github.com/travboz/greenlightv3/internal/data"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or
			// not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")

				// The value returned by recover() has the type any, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using
				// our custom Logger type at the ERROR level and send the client a 500
				// Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Rate limit by the server as a whole.
// func (app *application) globalRateLimit(next http.Handler) http.Handler {
// 	// Initialize a new rate limiter which allows an average of 2 requests per second,
// 	// with a maximum of 4 requests in a single ‘burst’.
// 	limiter := rate.NewLimiter(2, 4)

// 	// The function we are returning is a closure, which 'closes over' the limiter
// 	// variable.
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Call limiter.Allow() to see if the request is permitted, and if it's not,
// 		// then we call the rateLimitExceededResponse() helper to return a 429 Too Many
// 		// Requests response (we will create this helper in a minute).
// 		if !limiter.Allow() {
// 			app.rateLimitExceededResponse(w, r)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// Rate limit by client IP.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Define a client struct to hold the rate limiter and last seen time for each
	// client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
		// ClientCleanupTime = 3 * time.Minute
	)

	// Launch a background goroutine which removes old entries from the clients map once
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Importantly, unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Only carry out the check if rate limiting is enabled.
		if app.config.limiter.enabled {
			// Use the realip.FromRequest() function to get the client's real IP address.
			ip := realip.FromRequest(r)

			// Lock the mutex to prevent this code from being executed concurrently.
			mu.Lock()

			// Check to see if the IP address already exists in the map. If it doesn't, then
			// initialize a new rate limiter and add the IP address and limiter to the map.
			if _, found := clients[ip]; !found {
				// Create and add a new client struct to the map if it doesn't already exist.
				clients[ip] = &client{
					limiter: rate.NewLimiter(
						rate.Limit(app.config.limiter.rps),
						app.config.limiter.burst),
				}
			}

			// Update the last seen time for the client.
			clients[ip].lastSeen = time.Now()

			// Case: not allowed to make another request
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Case: haven't reached their limit, can make more requests
			// Very importantly, unlock the mutex before calling the next handler in the
			// chain. Notice that we DON'T use defer to unlock the mutex, as that would mean
			// that the mutex isn't unlocked until all the handlers downstream of this
			// middleware have also returned.
			mu.Unlock()

		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// func (app *application) authenticate(next http.Handler) http.Handler {
// 	fn := func(w http.ResponseWriter, r *http.Request) {
// 		// Add the "Vary: Authorization" header to the response. This indicates to any
// 		// caches that the response may vary based on the value of the Authorization
// 		// header in the request.
// 		w.Header().Add("Vary", "Authorization")

// 		// Retrieve the value of the Authorization header from the request. This will
// 		// return the empty string "" if there is no such header found.
// 		authorizationHeader := r.Header.Get("Authorization")

// 		// If there is no Authorization header found, use the contextSetUser() helper
// 		// that we just made to add the AnonymousUser to the request context. Then we
// 		// call the next handler in the chain and return without executing any of the
// 		// code below.
// 		if authorizationHeader == "" {
// 			r = app.contextSetUser(r, data.AnonymousUser)
// 			next.ServeHTTP(w, r)
// 			return
// 		}

// 		// Otherwise, we expect the value of the Authorization header to be in the format
// 		// "Bearer <token>". We try to split this into its constituent parts, and if the
// 		// header isn't in the expected format we return a 401 Unauthorized response
// 		// using the invalidAuthenticationTokenResponse() helper (which we will create
// 		// in a moment).
// 		headerParts := strings.Split(authorizationHeader, " ")
// 		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
// 			app.invalidAuthenticationTokenResponse(w, r)
// 			return
// 		}

// 		// Extract the actual authentication token from the headee parts
// 		token := headerParts[1]

// 		// Validate the token to make sure it is in a sensible format.
// 		v := validator.New()

// 		// If the token isn't valid, use the invalidAuthenticationTokenResponse()
// 		// helper to send a response, rather than the failedValidationResponse() helper
// 		// that we'd normally use.
// 		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
// 			app.invalidAuthenticationTokenResponse(w, r)
// 			return
// 		}

// 		// Retrieve the details of the user associated with the authentication token,
// 		// again calling the invalidAuthenticationTokenResponse() helper if no
// 		// matching record was found. IMPORTANT: Notice that we are using
// 		// ScopeAuthentication as the first parameter here.
// 		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
// 		if err != nil {
// 			switch {
// 			case errors.Is(err, data.ErrRecordNotFound):
// 				app.invalidAuthenticationTokenResponse(w, r)
// 			default:
// 				app.serverErrorResponse(w, r, err)
// 			}

// 			return
// 		}

// 		// Call the contextSetUser() helper to add the user information to the request
// 		// context.
// 		r = app.contextSetUser(r, user)

// 		// Call the next handler in the chain.
// 		next.ServeHTTP(w, r)
// 	}

// 	return http.HandlerFunc(fn)
// }

func (app *application) authenticateJWT(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>".
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Extract the actual authentication token from the headee parts
		token := headerParts[1]

		// Parse and extract claims.
		// Returns an error if JWT contents don't match the signature/secret key.
		claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secret))
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Check if the JWT is still valid at this moment in time.
		if !claims.Valid(time.Now()) {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Check that the issuer is our application.
		if claims.Issuer != "greenlight.travboz.net" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Check that our application is in the expected audiences for the JWT.
		if !claims.AcceptAudience("greenlight.travboz.net") {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// At this point, we know that the JWT is all OK and we can trust the data in
		// it. We extract the user ID from the claims subject and convert it from a
		// string into an int64.
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Lookup the user record from the database.
		user, err := app.models.Users.GetById(userID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Call the contextSetUser() helper to add the user information to the request
		// context.
		r = app.contextSetUser(r, user)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
// 	fn := func(w http.ResponseWriter, r *http.Request) {
// 		user := app.contextGetUser(r)

// 		// If the user is anonymous, then call the authenticationRequiredResponse() to
// 		// inform the client that they should authenticate before trying again.
// 		if user.IsAnonymous() {
// 			app.authenticationRequiredResponse(w, r)
// 			return
// 		}

// 		// If the user is not activated, use the inactiveAccountResponse() helper to
// 		// inform them that they need to activate their account.
// 		if !user.Activated {
// 			app.inactiveAccountResponse(w, r)
// 			return
// 		}

// 		// Call the next handler in the chain.
// 		next.ServeHTTP(w, r)
// 	}

// 	return http.HandlerFunc(fn)
// }

// Create a new requireAuthenticatedUser() middleware to check that a user is not
// anonymous.
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Checks that a user is both authenticated and activated.
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		// Check that the user is activated
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	// Wrap fn with the requireAuthenticatedUser() middleware before returning it.
	return app.requireAuthenticatedUser(fn)

	// The flow:
	// Call requireAuthenticatedUser to CHECK if the user is authenticated, then
	// Call requireActivatedUser to CHECK the user is activated.
	// Order:
	// 1st requiredAuthenticatedUSer
	// 2nd requireActivatedUser

	// 1st runs then our 2nd middleare does.
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user from the request context.
		user := app.contextGetUser(r)

		// Fetch the permissions the user has
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Check if they have the required permission. If they don't, then
		// return a 403 Forbidden response.
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// Otherwise they have the required permission so we call the next handler in
		// the chain.
		next.ServeHTTP(w, r)
	}

	// Wrap this with the requireActivatedUser() middleware before returning it.
	// So we call requireAuthenticatedUser -> requireActivatedUser -> finally requirePermissionCode
	return app.requireActivatedUser(fn)
}

// Simple CORS requests
// // Enable ALL origins
// func (app *application) enableCORS(next http.Handler) http.Handler {
// 	fn := func(w http.ResponseWriter, r *http.Request) {
// 		// Add the "Vary: Origin" header.
// 		w.Header().Add("Vary", "Origin")

// 		// Get the value of the request's Origin header
// 		origin := r.Header.Get("Origin")

// 		// Only run this if there's an Origin request header present.
// 		if origin != "" {
// 			// Loop through the list of trusted origins, checking to see if the request
// 			// origin exactly matches one of them. If there are no trusted origins, then
// 			// the loop won't be iterated.
// 			for i := range app.config.cors.trustedOrigins {
// 				if origin == app.config.cors.trustedOrigins[i] {
// 					// If there is a match, then set a "Access-Control-Allow-Origin"
// 					// response header with the request origin as the value and break
// 					// out of the loop.
// 					w.Header().Set("Access-Control-Allow-Origin", origin)
// 					break
// 				}
// 			}
// 		}

// 		next.ServeHTTP(w, r)
// 	}

// 	return http.HandlerFunc(fn)
// }

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		// Add the "Vary: Access-Control-Request-Method" header
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Check if the request has the HTTP method OPTIONS and contains the
					// "Access-Control-Request-Method" header. If it does, then we treat
					// it as a preflight request.
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// Set the necessary preflight response headers, as discussed
						// previously.
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						// Write the headers along with a 200 OK status and return from
						// the middleware with no further action.
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// func (app *application) metrics(next http.Handler) http.Handler {
// 	// Initialize the new expvar variables when the middleware chain is first built.
// 	var (
// 		totalRequestsReceived           = expvar.NewInt("total_requests_received")
// 		totalResponsesSent              = expvar.NewInt("total_responses_sent")
// 		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
// 	)

// 	// The following code will be run for every request...
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Record the time that we started to process the request.
// 		start := time.Now()

// 		// Use the Add() method to increment the number of requests received by 1.
// 		totalRequestsReceived.Add(1)

// 		// Call the next handler in the chain.
// 		next.ServeHTTP(w, r)

// 		// On the way back up the middleware chain, increment the number of responses
// 		// sent by 1.
// 		totalResponsesSent.Add(1)

// 		// Calculate the number of microseconds since we began to process the request,
// 		// then increment the total processing time by this amount.
// 		duration := time.Since(start).Microseconds()
// 		totalProcessingTimeMicroseconds.Add(duration)
// 	})
// }

// The metricsResponseWriter type wraps an existing http.ResponseWriter and also
// contains a field for recording the response status code, and a boolean flag to
// indicate whether the response headers have already been written.
type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// This function returns a new metricsResponseWriter instance which wraps a given
// http.ResponseWriter and has a status code of 200 (which is the status
// code that Go will send in a HTTP response by default).
func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

// The Header() method is a simple 'pass through' to the Header() method of the
// wrapped http.ResponseWriter.
func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

// Again, the WriteHeader() method does a 'pass through' to the WriteHeader()
// method of the wrapped http.ResponseWriter. But after this returns,
// we also record the response status code (if it hasn't already been recorded)
// and set the headerWritten field to true to indicate that the HTTP response
// headers have now been written.
func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

// Likewise the Write() method does a 'pass through' to the Write() method of the
// wrapped http.ResponseWriter. Calling this will automatically write any
// response headers, so we set the headerWritten field to true.
func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	mw.headerWritten = true
	return mw.wrapped.Write(b)
}

// We also need an Unwrap() method which returns the existing wrapped
// http.ResponseWriter.
func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

func (app *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")

		// Declare a new expvar map to hold the count of responses for each HTTP status
		// code.
		totalResponsesSentByStatus = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Add(1)

		// Create a new metricsResponseWriter, which wraps the original
		// http.ResponseWriter value that the metrics middleware received.
		mw := newMetricsResponseWriter(w)

		// Call the next handler in the chain using the new metricsResponseWriter
		// as the http.ResponseWriter value.
		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)

		// At this point, the response status code should be stored in the
		// mw.statusCode field. Note that the expvar map is string-keyed, so we
		// need to use the strconv.Itoa() function to convert the status code
		// (which is an integer) to a string. Then we use the Add() method on
		// our new totalResponsesSentByStatus map to increment the count for the
		// given status code by 1.
		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
