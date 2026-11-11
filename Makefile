.PHONY: all build test parity-test benchmark seed run debug docker-up docker-down clean

all: build test

build:
	go build -v ./...

test:
	go test -v ./...

parity-test:
	go test -v ./internal/validation/... -run Test11RulesParityAgainstGroundTruth

benchmark:
	go test ./internal/validation/... -run '^$$' -bench BenchmarkValidationAllRecords -benchmem -count=1

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

seed:
	go run ./cmd/seed/main.go

run:
	go run ./cmd/api/main.go

debug:
	dlv debug ./cmd/api

clean:
	rm -f parity_test_results.txt
	go clean
