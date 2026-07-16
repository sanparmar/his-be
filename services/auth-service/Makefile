.PHONY: help build test proto-gen lint docker migrate-up migrate-down clean dev cover

BINARY_NAME=auth-service
SERVICE_NAME=his-auth-service
DOCKER_IMAGE=his-auth-service:latest
GOFLAGS=-v

help:
	@echo "Auth Service Makefile"
	@echo "===================="
	@echo "make build         - Build the service binary"
	@echo "make test          - Run all tests"
	@echo "make cover         - Run tests with coverage report"
	@echo "make proto-gen     - Generate gRPC code from .proto files"
	@echo "make lint          - Run golangci-lint"
	@echo "make docker        - Build Docker image"
	@echo "make migrate-up    - Run database migrations UP"
	@echo "make migrate-down  - Run database migrations DOWN"
	@echo "make clean         - Clean build artifacts"
	@echo "make dev           - Run service locally (requires postgres)"

build:
	@echo "Building $(BINARY_NAME)..."
	go build $(GOFLAGS) -o ./bin/$(BINARY_NAME) ./cmd/grpc
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
	protoc --proto_path=. --proto_path="C:\Tools\protoc\include" --proto_path="C:\Users\zeesh\go\pkg\mod\google.golang.org\genproto\googleapis\api@v0.0.0-20260715232425-e75dac1f907d" \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		proto/auth/v1/auth.proto
	cp proto/auth/v1/*.go api/auth/v1/
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
	docker run -p 9090:9090 -p 8080:8080 \
		-e DATABASE_URL=postgres://postgres:postgres@host.docker.internal:5432/auth_db?sslmode=disable \
		-e JWT_ACCESS_SECRET=access-secret-key \
		-e JWT_REFRESH_SECRET=refresh-secret-key \
		$(DOCKER_IMAGE)

migrate-up:
	@echo "Running migrations UP..."
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable" up
	@echo "✓ Migrations applied"

migrate-down:
	@echo "Running migrations DOWN..."
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable" down
	@echo "✓ Migrations rolled back"

dev: build
	@echo "Starting auth service..."
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable \
	JWT_ACCESS_SECRET=access-secret-key \
	JWT_REFRESH_SECRET=refresh-secret-key \
	GRPC_PORT=9090 \
	HTTP_PORT=8080 \
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

# Full CI pipeline
ci: proto-gen lint test
	@echo "✓ CI pipeline passed"