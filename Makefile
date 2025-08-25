include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

.PHONY: confirm help

## help: prints this help message
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# Create the new confirm target - add to anything destructive or dangerous.
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# generate_secret_key: generate a cryptographically secure random string with an underlying entropy of at least 32 bytes 
generate_secret_key: 
	openssl rand -hex 32


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

ENTRYPOINT_DIR=./cmd/api
OUTPUT_BINARY=./bin/greenlight
MIGRATIONS_PATH=./migrations

.PHONY: run/server/default
run/server/default:
	@echo "Starting server..."
	@go run ${ENTRYPOINT_DIR}

## run: run the application
.PHONY: run
run: run/server/default

## setup: setup and dependencies for the application
.PHONY: setup
# setup: compose/up db/migrations/up db/populate
setup: compose/up

.PHONY: run/binary
run/binary: build/api
	@echo "Starting API..."
	@${OUTPUT_BINARY}

.PHONY: db/container/run/postgres
db/container/run/postgres:
	@echo "Starting postgres database container..."
	@docker run --rm --name pgdb -e POSTGRES_PASSWORD=pa55word -d postgres:latest

.PHONY: postgres
postgres: db/container/run/postgres

## show/dsn: echo the database dsn for Greenlight
.PHONY: show/dsn
show/dsn:
	@echo ${GREENLIGHT_DB_DSN}

# docker commands

## compose/up: run docker compose up for all the services
.PHONY: compose/up
compose/up:
	@echo "Starting containers..."
	GIT_HASH=$$(git rev-parse HEAD) \
	GIT_DIRTY=$$(test -n "$$(git status --porcelain)" && echo true || echo false) \
	docker compose up -d --build

## compose/down: run docker compose down for all the services
.PHONY: compose/down
compose/down:
	@echo "Stopping containers..."
	@docker compose down -v

.PHONY: compose/restart
compose/restart: compose/down compose/up

## docker/container/connect/pgdb: connect to the postgres docker container and run a bash instance
.PHONY: docker/container/connect/pgdb
docker/container/connect/pgdb:
	@echo "Connecting to postgres container via shell.."
	@docker exec -it $(GREENLIGHT_DB_CONTAINER_NAME) /bin/bash

.PHONY: pgdb/validatelogs
pgdb/validatelogs:
	@echo "Retrieving logs from pgdb...""
	@docker logs -f $(GREENLIGHT_DB_CONTAINER_NAME)

## pgdb/connect: connect to the database using psql
.PHONY: pgdb/connect
pgdb/connect:
	@echo "Connecting to pgdb..."
	@docker exec -it $(GREENLIGHT_DB_CONTAINER_NAME) psql --username=greenlight --dbname=greenlight
.PHONY: pgdb/connect/independent
pgdb/connect/independent:
	@echo "Connecting to pgdb using port..."
	@psql --host=localhost --port=${DB_ACCESS_PORT} --dbname=greenlight --username=greenlight


# Migrations

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo "Creating migration files for ${name}..."
	@migrate create -seq -ext=.sql -dir=$(MIGRATIONS_PATH) ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "Running up migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(MIGRATIONS_DSN) up

## db/migrations/down: apply all down migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo "Running down migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(MIGRATIONS_DSN) down $(filter-out $@,$(MAKECMDGOALS))

# e.g. migrate/goto/version 1 -> rolls back to migration 1
.PHONY: db/migrations/goto/version
db/migrations/goto/version:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(MIGRATIONS_DSN) goto $(filter-out $@,$(MAKECMDGOALS))

# show current migration version
.PHONY: db/migrations/version
db/migrations/version:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(MIGRATIONS_DSN) version

# Used for cleaning a dirty database.
# 1st: Manually roll back partial changes to DB - i.e. fix errors in migration in question. 
# 2nd: Run the below rule with the DB version you want. # eg: migrate -path=./migrations -database=$EXAMPLE_DSN force 1
.PHONY: db/migrations/force
db/migrations/force:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(MIGRATIONS_DSN) force ${version}

.PHONY: db/populate
db/populate:
	@echo "Seeding database..."
	# todo
	@echo "Done."

BASE_URL=http://localhost:4000

# Testing rate limiter
.PHONY: test/ratelimiter
test/ratelimiter:
	@for i in {1..6}; do curl ${BASE_URL}/v1/healthcheck; done

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo "Verifying and vendoring module dependencies..."
	go mod verify
	go mod vendor

# Vendoring dependencies in this way basically stores a complete copy of the source code for third-party packages in a vendor folder in your project.
# The `go mod vendor` command will then copy the necessary source code from your module cache into a new vendor directory in your project root.

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

TARGET_OS=linux
TARGET_ARCH=amd64

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building binary...'
	@go build -o=${OUTPUT_BINARY} ${ENTRYPOINT_DIR}
	GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go build -ldflags='-s' -o=./bin/${TARGET_OS}_${TARGET_ARCH}/${OUTPUT_BINARY} ${ENTRYPOINT_DIR}