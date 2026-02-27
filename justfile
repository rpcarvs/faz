default:
    just --list

# Build the project
build:
    go build .

# Run tests
test:
    go test . -v

# Run tests with coverage
coverage:
    go test . -coverprofile=coverage.out
    go tool cover -html=coverage.out

# Format code
fmt:
    go fmt .

# Lint using go vet (built-in)
lint:
    go vet .
    golangci-lint run

# Update dependencies
tidy:
    go mod tidy

ci: fmt lint tidy

# Clean build artifacts
clean:
    rm -rf bin coverage.out
