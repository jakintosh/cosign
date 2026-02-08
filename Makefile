.DEFAULT_GOAL := build

.PHONY: generate build run init test install clean

generate:
	go generate ./cmd/cosign

build:
	mkdir -p bin
	go build -o bin/cosign ./cmd/cosign

run: build
	mkdir -p ./data
	./bin/cosign serve --db-path ./data/cosign.db --credentials-directory ./secrets/

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
