BINARY := bin/appointment-service

.PHONY: build test run clean

build:
	go build -o $(BINARY) ./cmd/server

test:
	go test ./...

run:
	go run ./cmd/server

clean:
	rm -f $(BINARY)
