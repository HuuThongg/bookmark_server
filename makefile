run:
	go run cd/api/main.go

tidy:
	go mod tidy

sqlc:
	sqlc generate

compile:
	sqlc compile
