# Start from the official Golang image
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
RUN go build -o server .

# Use a minimal image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder
COPY --from=builder /app/server .

# Expose the port your app runs on
EXPOSE 8080

# Run the binary
CMD ["./server"]