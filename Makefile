.PHONY: generate build test install clean

generate:
	go generate ./cmd/cosign

build:
	mkdir -p bin
	go build -o bin/cosign ./cmd/cosign

test:
	go test ./...

install:
	go install ./cmd/cosign

clean:
	rm -rf bin/
