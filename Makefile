ENTRYPOINT_DIR=./cmd/api

.PHONY: run/server/default
run/server/default:
	@echo "Starting server..."
	@go run ${ENTRYPOINT_DIR}

.PHONY: run
run: run/server/default


ENTRYPOINT_DIR=./cmd/api
OUTPUT_BINARY=./bin/greenlight

.PHONY: build/api
build/api:
	@echo 'Building into binary...'
	go build -o=${OUTPUT_BINARY} ${ENTRYPOINT_DIR}