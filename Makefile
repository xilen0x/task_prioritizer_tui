.PHONY: build run demo test vet fmt check update-golden

build:
	go build -o bin/task_prioritizer ./cmd/task_prioritizer

run:
	go run ./cmd/task_prioritizer

demo:
	go run ./cmd/task_prioritizer --demo

test:
	go test ./...

vet:
	go vet ./...

fmt:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

check: fmt vet test

update-golden:
	go test ./internal/tui/... -update
