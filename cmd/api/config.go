package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
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

func LoadEnv() error {
	return godotenv.Load()
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

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	// Read the connection pool settings from command-line flags into the config struct.
	// Notice that the default values we're using are the ones we discussed above?
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "ce6c4f9b850da4", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "49364c7bb5284d", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Trav <travis.collab@gmail.com>", "SMTP sender")

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
	flag.StringVar(&cfg.jwt.secret, "jwt-secret", "", "JWT secret")

	// Create a new version boolean flag with the default value of false.
	flag.BoolVar(&cfg.displayVersion, "version", false, "Display version and exit")

	flag.Parse()

	return cfg
}
