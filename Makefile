.PHONY: help test up

up:
	go run cmd/api/main.go

test:
	go test ./... -v

help:
	@echo "Available commands:"
	@echo "  up   - Run the main.go in development mode"
	@echo "  test - Run all tests with verbose output"
