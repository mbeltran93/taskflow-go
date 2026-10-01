.PHONY: build test test-unit test-integration vet fmt run compose-up compose-down

build:
	go build -o bin/taskflow-go ./cmd/api

test-unit:
	go test ./internal/...

test-integration:
	go test ./test/...

test: test-unit test-integration

vet:
	go vet ./...

fmt:
	gofmt -l .

run:
	go run ./cmd/api

compose-up:
	docker compose up --build

compose-down:
	docker compose down -v
