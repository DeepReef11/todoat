.PHONY: build test clean run docker-up docker-down docker-wait test-integration test-nextcloud test-todoist

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

# Docker targets for integration testing
# Start Docker containers for integration tests (Nextcloud)
docker-up:
	docker compose up -d
	@$(MAKE) docker-wait

# Wait for Docker containers to be healthy
docker-wait:
	@echo "Waiting for Nextcloud to be healthy..."
	@until docker inspect --format='{{.State.Health.Status}}' todoat-nextcloud-test 2>/dev/null | grep -q healthy; do \
		echo "Waiting for Nextcloud..."; \
		sleep 5; \
	done
	@echo "Nextcloud is healthy!"

# Stop and remove Docker containers
docker-down:
	docker compose down -v

# Run all integration tests (requires Docker containers running)
test-integration: docker-up
	TODOAT_NEXTCLOUD_HOST=localhost:8080 \
	TODOAT_NEXTCLOUD_USERNAME=admin \
	TODOAT_NEXTCLOUD_PASSWORD=adminpass \
	go test -tags=integration -v ./...

# Run Nextcloud integration tests only
test-nextcloud: docker-up
	TODOAT_NEXTCLOUD_HOST=localhost:8080 \
	TODOAT_NEXTCLOUD_USERNAME=admin \
	TODOAT_NEXTCLOUD_PASSWORD=adminpass \
	go test -tags=integration -v -run Integration ./backend/nextcloud

# Run Todoist integration tests only (requires TODOAT_TODOIST_TOKEN env var)
test-todoist:
	go test -tags=integration -v -run Integration ./backend/todoist
