# Use a smaller Go base image
FROM golang:1.24.1-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the .env.local file
COPY .env.local .env.local

# Copy the go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the entire source code
COPY . .

# Build the application for production
RUN GOOS=linux GOARCH=amd64 go build -o server ./cmd

# Create a smaller final image
FROM alpine:latest

# Install necessary dependencies for the application to run
RUN apk add --no-cache ca-certificates

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server /app/

# Expose the port
EXPOSE 8081

# Run the server
CMD ["./server"]