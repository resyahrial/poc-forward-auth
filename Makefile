.PHONY: help build up down restart logs test clean admin user guest no-auth access-logs dashboard

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build all services
	docker-compose build

up: ## Start all services
	docker-compose up -d
	@echo "Services are starting..."
	@echo "Wait a few seconds for services to be ready"
	@echo ""
	@echo "Traefik Dashboard: http://localhost:8080"
	@echo "Book Service (via Traefik): http://localhost/books"
	@echo "Auth Service (direct): http://localhost:8081"
	@echo ""
	@echo "Run 'make test' to test the setup"

down: ## Stop all services
	docker-compose down

restart: down up ## Restart all services

logs: ## Show logs from all services
	docker-compose logs -f

logs-auth: ## Show auth service logs
	docker-compose logs -f auth-service

logs-book: ## Show book service logs
	docker-compose logs -f book-service

logs-traefik: ## Show Traefik logs
	docker-compose logs -f traefik

test: ## Run integration tests
	@echo "Running integration tests..."
	@./test.sh

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@./run-tests.sh

test-auth: ## Run auth service unit tests
	@echo "Testing auth service..."
	@cd auth-service && go test -v -cover

test-book: ## Run book service unit tests
	@echo "Testing book service..."
	@cd book-service && go test -v -cover

test-all: test-unit test ## Run all tests (unit + integration)

admin: ## Test with admin token (sees all books)
	@echo "Testing with admin token..."
	@curl -s http://localhost/books \
		-H "Authorization: Bearer admin:admin:secret-token" | jq

user: ## Test with user token (sees public + user books)
	@echo "Testing with user token..."
	@curl -s http://localhost/books \
		-H "Authorization: Bearer alice:user:secret-token" | jq

guest: ## Test with guest token (sees public books only)
	@echo "Testing with guest token..."
	@curl -s http://localhost/books \
		-H "Authorization: Bearer bob:guest:secret-token" | jq

no-auth: ## Test without authentication (should fail)
	@echo "Testing without authentication..."
	@curl -v http://localhost/books

access-logs: ## View access logs
	@echo "Fetching access logs..."
	@curl -s http://localhost/access-logs | jq

dashboard: ## Open Traefik dashboard
	@echo "Opening Traefik dashboard..."
	@open http://localhost:8080 || xdg-open http://localhost:8080 || echo "Please open http://localhost:8080 in your browser"

clean: ## Remove all containers, images, and volumes
	docker-compose down -v
	docker system prune -f

status: ## Check status of services
	@docker-compose ps
