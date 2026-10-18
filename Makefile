.PHONY: build run demo test vet fmt check update-golden

build:
	go build -o bin/tareas ./cmd/tareas

run:
	go run ./cmd/tareas

demo:
	go run ./cmd/tareas --demo

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

check: fmt vet test

update-golden:
	go test ./internal/tui/... -update
