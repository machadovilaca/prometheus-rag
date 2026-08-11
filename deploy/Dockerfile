# Production Dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o prometheus-rag ./main.go

# Final stage - use distroless for security
FROM gcr.io/distroless/base-debian11:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/prometheus-rag /prometheus-rag

# Use non-root user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/prometheus-rag", "healthcheck"] || exit 1

ENTRYPOINT ["/prometheus-rag"]
