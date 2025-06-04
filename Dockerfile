# Build stage
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o csv-importer ./cmd/csv-importer

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/csv-importer .

# Expose the application port
EXPOSE 8080

# Set environment variables
ENV TZ=UTC

# Run the application
CMD ["./csv-importer"] 