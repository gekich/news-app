################################################################################
# Build Commands
################################################################################

up: ## Start the application
	docker-compose up -d

down: ## Stop and remove all running containers
	docker-compose down

remove: ## Stop the application, remove containers and delete all associated volumes
	docker-compose down -v

################################################################################
# Test & Quality tools
################################################################################

test-unit: ## Run unit tests
	@go test -v -tags=unit ./...

test-integration: ## Run integration tests
	@go test -v -tags=integration ./...

test-all: ## Run all test including unit and integration
	@go test -v -tags="unit,integration" ./...

coverage: ## Generate and open coverage report
	@echo "Generating coverage reports..."
	@go test -tags=unit -coverprofile=unit.out ./...
	@go test -tags=integration -coverprofile=integration.out ./...
	@echo "mode: set" > coverage.out
	@tail -q -n +2 unit.out integration.out >> coverage.out
	@echo "Opening HTML coverage report..."
	@go tool cover -html=coverage.out

clean: ## Delete coverage report files
	@echo "Cleaning up..."
	@rm -f *.out

# ==============================================================================
# Help
# ==============================================================================

.PHONY: up down remove test-unit test-integration test-all coverage clean help
.DEFAULT_GOAL := help
help: ## List all commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)