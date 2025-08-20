# Justfile for RegexpTable helpers

# Default recipe lists available commands
default:
    @just --list

# Build all Go code in the helpers directory
build:
    go build ./...

# Run all tests with verbose output
test:
    go test -v ./...

# Run tests with coverage report
test-coverage:
    go test -v -cover ./...

# Run tests with detailed coverage report
test-coverage-html:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Clean generated files
clean:
    rm -f coverage.out coverage.html
    
# Run the builder example
run-example:
    cd example && go run main.go

# Format all Go code
fmt:
    go fmt ./...

# Run go vet on all packages
vet:
    go vet ./...

# Run golangci-lint
lint:
    golangci-lint run

# Run all quality checks (fmt, vet, lint, test)
check: fmt vet lint test

# Show Go module information
mod-info:
    go list -m all

# Tidy go.mod
mod-tidy:
    go mod tidy
