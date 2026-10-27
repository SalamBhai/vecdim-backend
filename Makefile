.PHONY: all build test parity-test seed run docker-up docker-down clean

all: build test

build:
	go build -v ./...

test:
	go test -v ./...

parity-test:
	go test -v ./internal/validation/... -run Test11RulesParityAgainstGroundTruth

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

seed:
	go run ./cmd/seed/main.go

run:
	go run ./cmd/api/main.go

clean:
	rm -f parity_test_results.txt
	go clean
