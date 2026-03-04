# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# We use tags !debug to ensure we use the pgx driver
RUN CGO_ENABLED=1 go build -tags "!debug" -o main ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy any other necessary files (like migrations if they aren't embedded, but they are)
# In this app, migrations are embedded using //go:embed migrations/*.sql

# Expose the port the app runs on
EXPOSE 9520

# Command to run the application
CMD ["./main"]
