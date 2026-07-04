.PHONY: help build test proto-gen lint docker migrate-up migrate-down clean dev cover

BINARY_NAME=patient-service
SERVICE_NAME=his-patient
DOCKER_IMAGE=his-patient:latest
GOFLAGS=-v

help:
	@echo "Patient Service Makefile"
	@echo "======================"
	@echo "make build         - Build the service binary"
	@echo "make test          - Run all tests"
	@echo "make cover         - Run tests with coverage report"
	@echo "make proto-gen     - Generate gRPC code from .proto files"
	@echo "make lint          - Run golangci-lint"
	@echo "make docker        - Build Docker image"
	@echo "make migrate-up    - Run database migrations UP"
	@echo "make migrate-down  - Run database migrations DOWN"
	@echo "make clean         - Clean build artifacts"
	@echo "make dev           - Run service locally (requires postgres + kafka)"

build:
	@echo "Building $(BINARY_NAME)..."
	go build $(GOFLAGS) -o ./bin/$(BINARY_NAME) ./cmd/server
	@echo "✓ Binary: ./bin/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	go test -v -race -count=1 ./...
	@echo "✓ Tests passed"

cover:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total
	@echo "✓ Coverage report: coverage.html"

proto-gen:
	@echo "Generating gRPC code from .proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/patient/v1/patient.proto
	@echo "✓ Proto code generated"

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...
	@echo "✓ Lint passed"

docker:
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	docker build -t $(DOCKER_IMAGE) .
	@echo "✓ Docker image built: $(DOCKER_IMAGE)"

docker-run:
	@echo "Running Docker container..."
	docker run -p 50051:50051 -p 8081:8081 \
		-e DB_HOST=host.docker.internal \
		-e DB_USER=postgres \
		-e DB_PASSWORD=postgres \
		-e DB_NAME=his_patient \
		-e KAFKA_BROKERS=host.docker.internal:9092 \
		$(DOCKER_IMAGE)

migrate-up:
	@echo "Running migrations UP..."
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/his_patient?sslmode=disable" up
	@echo "✓ Migrations applied"

migrate-down:
	@echo "Running migrations DOWN..."
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/his_patient?sslmode=disable" down
	@echo "✓ Migrations rolled back"

dev: build
	@echo "Starting patient service..."
	DB_HOST=localhost DB_USER=postgres DB_PASSWORD=postgres DB_NAME=his_patient \
	KAFKA_BROKERS=localhost:9092 \
	./bin/$(BINARY_NAME)

clean:
	@echo "Cleaning build artifacts..."
	rm -rf ./bin
	rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

run-tests-watch:
	@echo "Running tests in watch mode..."
	go test -v -race -count=1 ./... --watch

# Integration test with containers (requires Docker)
test-integration:
	@echo "Running integration tests with containers..."
	go test -v -race -tags=integration ./...
	@echo "✓ Integration tests passed"
