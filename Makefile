.PHONY: build test clean run

# Build the binary
build:
	go build -o bin/todoat ./cmd/todoat

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Remove build artifacts
clean:
	rm -rf bin/
	go clean

# Build and run with arguments
run: build
	./bin/todoat $(ARGS)
