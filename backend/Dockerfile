# Use the official Golang image
FROM golang:1.22-alpine AS builder

# Set the working directory
WORKDIR /app

#RUN apk add --update curl && \
#    rm -rf /var/cache/apk/*

#RUN curl -L -O https://artifacts.elastic.co/downloads/beats/elastic-agent/elastic-agent-8.15.2-linux-x86_64.tar.gz
#RUN tar xzvf elastic-agent-8.15.2-linux-x86_64.tar.gz
#RUN cd elastic-agent-8.15.2-linux-x86_64

#RUN ./elastic-agent install

# Copy go.mod and go.sum first to cache dependency layers
COPY go.mod go.sum ./

# Download dependencies (including tidying them)
RUN go mod tidy
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go binary
RUN go build -o main .

# Expose the port for the application
EXPOSE 8080

# Command to run the binary
CMD ["./main"]

