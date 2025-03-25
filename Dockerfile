# Build stage (minimal dependencies)
FROM golang:1.24.1-alpine AS builder

# Set CGO_ENABLED=0 for a fully static binary
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the source code
COPY . .

# Build the application statically
RUN go build -o server ./cmd

# Final minimal image
FROM scratch

# Set working directory
WORKDIR /app

# Copy built binary from the builder stage
COPY --from=builder /app/server /app/server

# Expose port
EXPOSE 8081

# Run the server
CMD ["/app/server"]
