# Build stage
FROM golang:1.24.3-alpine AS builder

# Install required dependencies (GCC + WebP)
RUN apk add --no-cache gcc musl-dev libwebp-dev

WORKDIR /app

# Copy dependencies and download them
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build with CGO enabled (required for chai2010/webp)
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o server ./cmd

# Final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates libwebp

WORKDIR /app
COPY --from=builder /app/server /app/

EXPOSE 8081
CMD ["/app/server"]
