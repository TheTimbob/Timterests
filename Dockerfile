# Stage 1: Build
FROM golang:latest AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency resolution
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application
RUN GOOS=linux go build -o main ./cmd/web

# Stage 2: Run
FROM ubuntu:22.04

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder
COPY --from=builder /app/main /app/main

# Copy static assets, templates, etc.
COPY --from=builder /app/internal/templates /app/internal/templates
COPY --from=builder /app/static /app/static

# Install runtime dependencies (e.g., database clients, TLS certs)
RUN apt-get update && apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]

