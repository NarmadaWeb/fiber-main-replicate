# Stage 1: Build stage
# Use an official Golang runtime as a parent image
FROM golang:1.24-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files to leverage Docker cache
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
RUN go mod verify

# Copy the source code into the container
COPY . .

# Build the main server application binary
# Disable CGO for static linking, specify target OS
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/main_server ./cmd/main_server/main.go

# Build the replica server application binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/replica_server ./cmd/replica_server/main.go

# Stage 2: Final stage/image
# Use a minimal alpine image for a small footprint
FROM alpine:latest

# Add ca-certificates to make https calls possible
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the configuration file from the current directory
# Ensure config.yml is present in the directory where `docker build` is run
COPY config.yml ./config.yml

# Copy the built binaries from the builder stage
COPY --from=builder /app/main_server /app/main_server
COPY --from=builder /app/replica_server /app/replica_server

# Expose ports (assuming default ports 8000 for main, 8001 for replica from config defaults)
# These are informational; you still need to map them with `docker run -p`
EXPOSE 8000
EXPOSE 8001

# Note: This Dockerfile builds both binaries.
# To run the main server: docker run -p 8000:8000 <image_name> /app/main_server
# To run the replica server: docker run -p 8001:8001 <image_name> /app/replica_server
# Consider using Docker Compose for managing both containers and potentially a database.
# The default CMD is not set, as you need to specify which binary to run when starting the container.
# For example, to run the main server by default, uncomment the following line:
# CMD ["/app/main_server"]
