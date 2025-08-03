ENTRYPOINT_DIR=./cmd/api

.PHONY: run/server/default
run/server/default:
	@echo "Starting server..."
	@go run ${ENTRYPOINT_DIR}