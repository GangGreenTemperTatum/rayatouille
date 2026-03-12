BINARY_NAME=rayatouille
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build test test-e2e clean lint fmt check

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/rayatouille

test:
	go test -race ./...

test-e2e:
	RAY_DASHBOARD_URL=http://localhost:8265 go test -race -tags=e2e ./test/e2e/...

clean:
	rm -rf bin/ dist/

fmt:
	gofmt -w .

lint:
	go vet ./...
	golangci-lint run

check: fmt lint test
	@echo "All checks passed"
