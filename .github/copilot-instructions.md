# Copilot Instructions for compose2podman

## Project Overview

compose2podman is a Go tool for converting Docker Compose files to Podman configurations.

## Go Environment

- **Go Version**: 1.25.7 (specified in go.mod)
- **Module Path**: `github.com/kad/compose2podman`

## Build and Test Commands

```bash
# Build the project
go build -o compose2podman ./cmd/compose2podman

# Run tests
go test ./...

# Run tests for a specific package
go test ./pkg/converter

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestFunctionName ./pkg/path

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Run linter (if golangci-lint is configured)
golangci-lint run

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Project Structure

As the project develops, follow this structure:

- `cmd/compose2podman/` - Main application entry point
- `pkg/` - Reusable packages
  - `pkg/converter/` - Core conversion logic from Compose to Podman
  - `pkg/parser/` - Docker Compose file parsing
  - `pkg/generator/` - Podman configuration generation
- `internal/` - Internal packages not meant for external use
- `testdata/` - Test fixtures and sample files

## Key Conventions

### Error Handling

- Use explicit error returns, not panics for business logic
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Log errors at the point of handling, not at every return

### Testing

- Place test files alongside the code they test (`file.go` → `file_test.go`)
- Use table-driven tests for testing multiple scenarios
- Store test fixtures in `testdata/` directories

## Code Organization

- Keep CLI logic separate from core conversion logic
- Make converter functions testable by avoiding direct file I/O in core logic
- Use interfaces for external dependencies (file system, network) to enable mocking

## Go Standards and Tooling

### Formatting

- **Always** run `gofmt` or `goimports` before committing
- Use `goimports` to automatically manage import statements
- Line length: no hard limit, but aim for readability (typically < 120 chars)
- Use tabs for indentation (Go standard)

### Static Analysis Tools

```bash
# Basic vet (catches common mistakes)
go vet ./...

# Enhanced linting with golangci-lint (recommended)
golangci-lint run

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Static analysis with staticcheck
staticcheck ./...
```

### Recommended golangci-lint Configuration

Create `.golangci.yml` with:
- Enable: `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`
- Enable: `gofmt`, `goimports`, `misspell`, `revive`
- Consider: `gocritic`, `gosec`, `dupl`

### Code Quality Practices

- Keep functions small and focused (< 50 lines when possible)
- Prefer explicit over clever code
- Document exported functions, types, and packages with godoc comments
- Start comments with the name of the thing being described
- Use meaningful variable names; avoid single letters except for short-lived loops

## CI/CD Practices

### GitHub Actions Workflow

Create `.github/workflows/ci.yml`:

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    strategy:
      matrix:
        go-version: [1.24.x, 1.25.x]
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - run: go mod download
      - run: go build -v ./...
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go vet ./...
```

### CI Checklist

- Run tests on multiple Go versions (current and previous)
- Test on multiple OS platforms (Linux, macOS, Windows)
- Enable race detector in CI (`-race`)
- Check test coverage (aim for > 80% for critical paths)
- Run `go vet` and linters
- Verify `go.mod` and `go.sum` are up to date

## Git Practices

### Commit Messages

- Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- First line: short summary (< 72 chars)
- Imperative mood: "Add feature" not "Added feature"

### Branch Strategy

- `main` - stable, production-ready code
- `develop` - integration branch for features
- Feature branches: `feature/descriptive-name`
- Bug fixes: `fix/issue-description`

### What to Commit

- **Always commit**: `go.mod`, `go.sum`
- **Never commit**: binaries, `vendor/` (unless vendoring), IDE configs (`.idea/`, `.vscode/`)
- **Consider**: `.gitignore` for Go projects

### Before Committing

```bash
go mod tidy                  # Clean up dependencies
go fmt ./...                 # Format code
go vet ./...                 # Check for issues
go test ./...                # Verify tests pass
golangci-lint run            # Run linters
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/sh
go fmt ./...
go vet ./...
go test ./...
```

## Docker Compose to Podman Considerations

When working on conversion logic:

- Podman uses different networking models than Docker Compose (CNI vs bridge)
- Volume mounting syntax differs between docker-compose and podman play kube
- Some Compose v3 features may not have direct Podman equivalents
- Consider both `podman-compose` and `podman play kube` as target formats
- Handle version differences in Docker Compose schema (v2, v3, v3.x)
