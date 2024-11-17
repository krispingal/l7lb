# syntax=docker/dockerfile:1
# Build stage
FROM golang:alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and go sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o loadbalancer ./cmd/api/

# Final stage
FROM alpine:latest  

WORKDIR /app

# Copy config 
COPY config /app/config

COPY cert.pem key.pem ./
# Copy the binary from the previous stage
COPY --from=builder /app/loadbalancer .

# Expose the required port
EXPOSE 8443

# Command to run the application
CMD ["./loadbalancer"]