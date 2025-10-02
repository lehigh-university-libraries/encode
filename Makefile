.PHONY: build deps lint test

BINARY_NAME=encode

deps:
	go get .
	go mod tidy

build: deps
	go build -o $(BINARY_NAME) .

lint:
	go fmt ./...
	golangci-lint run

	@if command -v yq > /dev/null 2>&1; then \
		echo "Running yq validation on YAML files..."; \
		yq . **/*.yml > /dev/null; \
	else \
		echo "yq not found, skipping YAML validation"; \
	fi

test: build
	go test -v -race ./...
