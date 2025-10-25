.PHONY: build clean

build:
	mkdir -p bin
	go build -o bin/cosign ./cmd/cosign

clean:
	rm -rf bin/
