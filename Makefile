include .env

ENTRYPOINT_DIR=./cmd/api

.PHONY: run/server/default
run/server/default:
	@echo "Starting server..."
	@go run ${ENTRYPOINT_DIR}

.PHONY: run
run: run/server/default

OUTPUT_BINARY=./bin/greenlight

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