package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/travboz/greenlightv3/pkg/env"
)

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the
// application (development, staging, production, etc.). We will read in these
// configuration settings from command-line flags when the application starts.
// Add db struct field which holds configuration settings for our db connection pool.
type config struct {
	port           int
	env            string
	db             dbConfig
	limiter        limiterConfig
	smtp           smtpConfig
	cors           corsConfig
	jwt            jwtConfig
	displayVersion bool
}

type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

// Add a new limiter struct containing fields for the requests-per-second and burst
// values, and a boolean field which we can use to enable/disable rate limiting
// altogether.
type limiterConfig struct {
	rps     float64 // refills per second
	burst   int
	enabled bool
}

type smtpConfig struct {
	host     string
	port     int
	username string
	password string
	sender   string
}

type corsConfig struct {
	trustedOrigins []string
}

type jwtConfig struct {
	secret string // JWT signing secret
}

func NewConfig() config {
	var cfg config

	flag.IntVar(&cfg.port, "port", env.GetInt("GREENLIGHT_API_PORT", 4000), "API server port")
	flag.StringVar(&cfg.env, "env", env.GetString("ENV_STAGE", "development"), "Environment (development|staging|production)")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	flag.StringVar(
		&cfg.db.dsn,
		"db-dsn",
		env.GetString("GREENLIGHT_DB_DSN", dsn),
		"PostgreSQL DSN",
	)

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", env.GetInt("DB_MAX_OPEN_CONNS", 25), "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", env.GetInt("DB_MAX_IDLE_CONNS", 25), "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", env.GetDuration("DB_MAX_IDLE_TIME", 15*time.Minute), "PostgreSQL max connection idle time (e.g. 1h30m)")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", float64(env.GetInt("RATE_LIMITER_RPS", 2)), "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", env.GetInt("RATE_LIMITER_BURST", 4), "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", env.GetBool("RATE_LIMITER_ENABLED", false), "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", env.GetString("SMTP_HOST", "sandbox.smtp.mailtrap.io"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", env.GetInt("SMTP_PORT", 25), "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", env.GetString("SMTP_USERNAME", "ce6c4f9b850da4"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", env.GetString("SMTP_PASSWORD", "49364c7bb5284d"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", env.GetString("SMTP_SENDER", "Trav <travis.collab@gmail.com>"), "SMTP sender")

	// Use the flag.Func() function to process the -cors-trusted-origins command line
	// flag. In this we use the strings.Fields() function to split the flag value into a
	// slice based on whitespace characters and assign it to our config struct.
	// Importantly, if the -cors-trusted-origins flag is not present, contains the empty
	// string, or contains only whitespace, then strings.Fields() will return an empty
	// []string slice.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// Parse the JWT signing secret from the command-line-flag. Notice that we leave the
	// default value as the empty string if no flag is provided.
	flag.StringVar(
		&cfg.jwt.secret,
		"jwt-secret",
		env.GetString(
			"JWT_SECRET",
			"pei3einoh0Beem6uM6Ungohn2heiv5lah1ael4joopie5JaigeikoozaoTew2Eh6",
		),
		"JWT secret",
	)

	// Create a new version boolean flag with the default value of false.
	flag.BoolVar(&cfg.displayVersion, "version", env.GetBool("DISPLAY_VERSION", false), "Display version and exit")

	flag.Parse()

	return cfg
}
