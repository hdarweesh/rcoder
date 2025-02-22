# Use official Golang image to build the Go executor
FROM golang:1.22-alpine AS builder

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first to take advantage of Docker caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy && go mod download

# Copy the entire application
COPY . .

# Install dependencies (if needed)
RUN apk add --no-cache git

RUN go version

# Build the Go executable
RUN go build -o executor .

# Run the Go executor
CMD ["go", "run", "executor.go"]
