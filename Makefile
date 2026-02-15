.PHONY: build test clean install fmt vet lint run-kube run-quadlet release snapshot

# Build the binary
build:
	go build -o compose2podman ./cmd/compose2podman

# Build with version info
build-version:
	go build -ldflags="-X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o compose2podman ./cmd/compose2podman

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...
	go test -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

# Run tests with race detector
test-race:
	go test -race ./...

# Clean build artifacts
clean:
	rm -f compose2podman
	rm -f pod.yaml full-pod.yaml test-pod.yaml
	rm -rf quadlet-output quadlet-test
	rm -f coverage.txt coverage.html

# Install the binary
install:
	go install ./cmd/compose2podman

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy
	go mod verify

# Run example: generate Kubernetes YAML
run-kube:
	go run ./cmd/compose2podman -input testdata/simple.yaml -type kube -output test-pod.yaml
	@echo "Generated: test-pod.yaml"
	@cat test-pod.yaml

# Run example: generate Quadlet files
run-quadlet:
	go run ./cmd/compose2podman -input testdata/docker-compose.yaml -type quadlet -output quadlet-test
	@echo "Generated files in: quadlet-test/"
	@ls -la quadlet-test/

# Run all checks
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/compose2podman-linux-amd64 ./cmd/compose2podman
	GOOS=linux GOARCH=arm64 go build -o dist/compose2podman-linux-arm64 ./cmd/compose2podman
	GOOS=darwin GOARCH=amd64 go build -o dist/compose2podman-darwin-amd64 ./cmd/compose2podman
	GOOS=darwin GOARCH=arm64 go build -o dist/compose2podman-darwin-arm64 ./cmd/compose2podman
	GOOS=windows GOARCH=amd64 go build -o dist/compose2podman-windows-amd64.exe ./cmd/compose2podman

# GoReleaser: Create a release (requires goreleaser)
release:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser is not installed. Install it with:"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	goreleaser release --clean

# GoReleaser: Test release locally
snapshot:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser is not installed. Install it with:"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	goreleaser release --snapshot --clean --skip=publish

# GoReleaser: Check configuration
release-check:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser is not installed. Install it with:"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	goreleaser check
