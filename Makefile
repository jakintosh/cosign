.DEFAULT_GOAL := build

.PHONY: generate build run run-dashboard init test install clean

generate:
	go generate ./cmd/cosign

build:
	mkdir -p bin
	go build -o bin/cosign ./cmd/cosign

run: build
	mkdir -p ./data
	./bin/cosign serve --db-path ./data/cosign.db --credentials-directory ./secrets/

run-dashboard: build
	mkdir -p ./data
	@set -e; \
		./bin/cosign serve --db-path ./data/cosign.db --credentials-directory ./secrets/ & \
		api_pid=$$!; \
		./bin/cosign dashboard --api-base-url http://localhost:8080 --credentials-directory ./secrets/ & \
		dashboard_pid=$$!; \
		trap 'kill $$api_pid $$dashboard_pid 2>/dev/null || true; wait $$api_pid $$dashboard_pid 2>/dev/null || true' INT TERM EXIT; \
		wait $$api_pid $$dashboard_pid

init: build
	mkdir -p ./secrets
	@if [ -f ./secrets/api_key ]; then \
		echo "./secrets/api_key already exists; skipping bootstrap"; \
	else \
		./bin/cosign env create local --base-url http://localhost:8080 --bootstrap > ./secrets/api_key; \
	fi

test:
	go test ./...

install:
	go install ./cmd/cosign

clean:
	rm -rf bin/
