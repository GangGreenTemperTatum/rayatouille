set dotenv-load

default:
    @just --list

# Build the binary
build:
    go build -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o bin/rayatouille ./cmd/rayatouille

# Run all tests
test:
    go test -race ./...

# Run e2e tests against live cluster
test-e2e:
    RAY_DASHBOARD_URL=${RAY_DASHBOARD_URL:-http://localhost:8265} go test -race -tags=e2e ./test/e2e/...

# Format all Go files
fmt:
    gofmt -w .

# Lint (vet + golangci-lint)
lint:
    go vet ./...
    golangci-lint run

# Run all checks
check: fmt lint test

# Run the TUI against a cluster
run *ARGS:
    go run ./cmd/rayatouille {{ARGS}}

# Clean build artifacts
clean:
    rm -rf bin/ dist/
