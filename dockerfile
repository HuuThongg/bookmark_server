# Use the official Golang image as a build stage
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files for dependency resolution
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application code
COPY . .

# Copy config.env
# COPY config.env /app/config.env
# Build the Go application for Linux amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin ./cmd/api/main.go

# Use a minimal base image for the final build
FROM alpine:latest

# Install necessary packages (e.g., for PostgreSQL support)
#RUN apk --no-cache add ca-certificates

# Set the working directory in the final image
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin .
# Copy the config.env file from the builder stage
# COPY --from=builder /app/config.env .
# Expose the port that the application will run on
EXPOSE 8080

# Command to run the application
CMD ["./bin"]
