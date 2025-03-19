FROM golang:1.23.3

WORKDIR /app

# Copy files
COPY . .

# Install dependencies
RUN go mod tidy

# Build the application correctly
RUN go build -o server ./cmd

# Expose the port
EXPOSE 8080

# Run the server
CMD ["./server"]