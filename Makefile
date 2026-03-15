.PHONY: build test vet lint run clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/gdrive-cli ./cmd/gdrive-cli

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

run:
	go run ./cmd/gdrive-cli

clean:
	rm -rf bin/
