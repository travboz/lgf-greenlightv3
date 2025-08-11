# ==================================================================================== #
# HELPERS
# ==================================================================================== #

.PHONY: confirm help

# Create the new confirm target - add to anything destructive or dangerous.
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## help: prints this help message
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## gnerate_secret_key: generate a cryptographically secure random string with an underlying entropy of at least 32 bytes 
gnerate_secret_key: 
	openssl rand -hex 32


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

include .env

ENTRYPOINT_DIR=./cmd/api
OUTPUT_BINARY=./bin/greenlight
MIGRATIONS_PATH=./migrations

.PHONY: run/server/default
run/server/default:
	@echo "Starting server..."
	@go run ${ENTRYPOINT_DIR} -port=4000 -env=development -db-dsn=${GREENLIGHT_DB_DSN} -db-max-open-conns=50 -db-max-idle-conns=50 -db-max-idle-time=2h30m

.PHONY: run
run: run/server/default

.PHONY: run/binary
run/binary: build/api
	@echo "Starting API..."
	@${OUTPUT_BINARY} -port=4000 -env=development -db-dsn=${GREENLIGHT_DB_DSN} -db-max-open-conns=50 -db-max-idle-conns=50 -db-max-idle-time=2h30m

.PHONY: build/api
build/api:
	@echo 'Building into binary...'
	@go build -o=${OUTPUT_BINARY} ${ENTRYPOINT_DIR}

.PHONY: db/container/run/postgres
db/container/run/postgres:
	@echo "Starting postgres database container..."
	@docker run --rm --name pgdb -e POSTGRES_PASSWORD=pa55word -d postgres:latest

.PHONY: postgres
postgres: db/container/run/postgres

.PHONY: show/dsn
show/dsn:
	@echo ${GREENLIGHT_DB_DSN}

# docker commands
.PHONY: compose/up
compose/up:	
	@echo "Starting containers..."
	docker compose up -d --build

.PHONY: compose/down
compose/down:
	@echo "Stopping containers..."
	@docker compose down -v

.PHONY: docker/container/connect/pgdb
ddocker/container/connect/pgdb:
	@echo "Connecting to postgres container via shell.."
	@docker exec -it $(GREENLIGHT_DB_CONTAINER_NAME) /bin/bash

.PHONY: pgdb/validatelogs
pgdb/validatelogs:
	@echo "Retrieving logs from pgdb...""
	@docker logs -f $(GREENLIGHT_DB_CONTAINER_NAME)

.PHONY: pgdb/connect
pgdb/connect:
	@echo "Connecting to pgdb..."
	@docker exec -it $(GREENLIGHT_DB_CONTAINER_NAME) psql --username=greenlight --dbname=greenlight

pgdb/connect/independent:
	@echo "Connecting to pgdb using port..."
	@psql --host=localhost --port=${DB_ACCESS_PORT} --dbname=greenlight --username=greenlight



# Migrations
.PHONY: migrate/create migrate/up migrate/down migrate/version migrate/force

db/migrate/create:
	@migrate create -seq -ext=.sql -dir=$(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

## db/migrations/new name=$1: create a new database migration
db/migrate/new:
	@echo "Creating migration files for ${name}..."
	@migrate create -seq -ext=.sql -dir=$(MIGRATIONS_PATH) ${name}

## db/migrate/up: apply all up database migrations
db/migrate/up: confirm
	@echo "Running up migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(GREENLIGHT_DB_DSN) up

db/migrate/down:
	@echo "Running down migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(GREENLIGHT_DB_DSN) down $(filter-out $@,$(MAKECMDGOALS))

# e.g. migrate/goto/version 1 -> rolls back to migration 1
db/migrate/goto/version:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(GREENLIGHT_DB_DSN) goto $(filter-out $@,$(MAKECMDGOALS))

db/migrate/version:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(GREENLIGHT_DB_DSN) version

# Used for cleaning a dirty database.
# 1st: Manually roll back partial changes to DB - i.e. fix errors in migration in question. 
# 2nd: Run the below rule with the DB version you want. # eg: migrate -path=./migrations -database=$EXAMPLE_DSN force 1
db/migrate/force:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(GREENLIGHT_DB_DSN) force $(filter-out $@,$(MAKECMDGOALS))

BASE_URL=http://localhost:4000

# Testing rate limiter
test/ratelimiter:
	@for i in {1..6}; do curl ${BASE_URL}/v1/healthcheck; done