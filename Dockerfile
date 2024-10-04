# Use the official Golang image as a build stage
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/app_prod ./cmd/api/main.go

FROM alpine:latest
# Install necessary packages (e.g., for PostgreSQL support)
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/app_prod .

EXPOSE 8080

CMD ["/app/app_prod"]
