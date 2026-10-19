.PHONY: build run demo test vet fmt check update-golden dist install uninstall

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GOARCH ?= $(shell go env GOARCH)
DIST_DIR := bin/dist
DIST_NAME := task_prioritizer_linux_$(GOARCH)

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

# Linux release tarball for $(GOARCH) (host arch by default), the layout install.sh expects.
dist:
	rm -rf $(DIST_DIR)/$(DIST_NAME)
	mkdir -p $(DIST_DIR)/$(DIST_NAME)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath \
		-ldflags "-s -w -X main.version=$(VERSION)" \
		-o $(DIST_DIR)/$(DIST_NAME)/task_prioritizer ./cmd/task_prioritizer
	cp packaging/linux/* LICENSE README.md $(DIST_DIR)/$(DIST_NAME)/
	tar -czf $(DIST_DIR)/$(DIST_NAME).tar.gz -C $(DIST_DIR) $(DIST_NAME)
	rm -rf $(DIST_DIR)/$(DIST_NAME)

# Install the local build under $(PREFIX) (default ~/.local) with a menu entry.
install: dist
	TASK_PRIORITIZER_TARBALL=$(DIST_DIR)/$(DIST_NAME).tar.gz sh install.sh

uninstall:
	sh install.sh --uninstall
