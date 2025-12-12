.PHONY: up

up:
	go run cmd/api/main.go

test:
	go test ./... -v
