FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/patient-service ./cmd/server

# Final stage - distroless
FROM gcr.io/distroless/base-debian11:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/patient-service /app/patient-service

# Copy config
COPY services/patient/configs/ /app/configs/

# Copy migrations
COPY services/patient/migrations/ /app/migrations/

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/patient-service", "-healthcheck"]

EXPOSE 50051 8081

USER nonroot:nonroot

ENTRYPOINT ["/app/patient-service"]
