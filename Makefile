.DEFAULT_GOAL := build

IMAGE_NAME ?= mattiasgees/spiffe-demo:latest

.PHONY: test
test:
	go test -v ./...
INIT_IMAGE_NAME ?= mattiasgees/spiffe-demo-init:latest
POSTGRES_IMAGE_NAME ?= mattiasgees/spiffe-postgres:latest
SPIFFE_GCP_PROXY_IMAGE_NAME ?= mattiasgees/spiffe-gcp-proxy:latest
export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: build # Build the container image
build:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(IMAGE_NAME) \
		--file Dockerfile \
		.
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(INIT_IMAGE_NAME) \
		./deploy/initcontainer
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(POSTGRES_IMAGE_NAME) \
		./deploy/postgresql
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(SPIFFE_GCP_PROXY_IMAGE_NAME) \
		./deploy/spiffe-gcp-proxy

.PHONY: publish # Push all the image to the remote registry
publish:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--output "type=image,push=true" \
		--tag $(IMAGE_NAME) \
		--file Dockerfile \
		.
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--output "type=image,push=true" \
		--tag $(INIT_IMAGE_NAME) \
		./deploy/initcontainer
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--output "type=image,push=true" \
		--tag $(POSTGRES_IMAGE_NAME) \
		./deploy/postgresql
	docker buildx build \
		--output "type=docker,push=true" \
		--tag $(SPIFFE_GCP_PROXY_IMAGE_NAME) \
		./deploy/spiffe-gcp-proxy

# ==================== Docker Compose Development Environment ====================
COMPOSE_DIR := deploy/docker-compose
COMPOSE := docker compose -f $(COMPOSE_DIR)/docker-compose.yaml

.PHONY: dev-up # Start the development environment with SPIRE and PostgreSQL
dev-up:
	@echo "Starting SPIFFE development environment..."
	$(COMPOSE) up -d
	@echo ""
	@echo "Development environment is starting..."
	@echo "  - SPIRE Server: managing workload identities"
	@echo "  - SPIRE Agent: providing workload API"
	@echo "  - PostgreSQL: database with certificate authentication"
	@echo "  - Ledger Service: https://localhost:8443"
	@echo ""
	@echo "Use 'make dev-logs' to view logs"
	@echo "Use 'make dev-down' to stop the environment"

.PHONY: dev-down # Stop and clean up the development environment
dev-down:
	@echo "Stopping development environment..."
	$(COMPOSE) down -v
	@echo "Development environment stopped and volumes removed."

.PHONY: dev-logs # View logs from all services
dev-logs:
	$(COMPOSE) logs -f

.PHONY: dev-status # Show status of all services
dev-status:
	$(COMPOSE) ps

.PHONY: dev-restart # Restart the development environment
dev-restart: dev-down dev-up

.PHONY: dev-build # Rebuild the ledger service and restart
dev-build:
	@echo "Rebuilding ledger service..."
	$(COMPOSE) build ledger
	$(COMPOSE) up -d ledger
	@echo "Ledger service rebuilt and restarted."

.PHONY: test-client # Start test client for manual testing
test-client:
	@echo "Starting test client..."
	$(COMPOSE) --profile test up -d test-client
	@echo ""
	@echo "Test client is running. Connect with:"
	@echo "  docker compose -f $(COMPOSE_DIR)/docker-compose.yaml exec test-client /bin/sh"
	@echo ""
	@echo "Example commands inside the container:"
	@echo "  # List accounts"
	@echo '  curl -k https://ledger:8443/api/accounts'
	@echo ""
	@echo "  # Create a transfer"
	@echo '  curl -k -X POST https://ledger:8443/api/transfers \'
	@echo '    -H "Content-Type: application/json" \'
	@echo '    -d '\''{"from_account":"11111111-1111-1111-1111-111111111111","to_account":"22222222-2222-2222-2222-222222222222","amount":"100.00"}'\'

.PHONY: dev-shell # Open a shell in the ledger container
dev-shell:
	$(COMPOSE) exec ledger /bin/sh

.PHONY: dev-psql # Connect to PostgreSQL
dev-psql:
	$(COMPOSE) exec postgres psql -U postgres -d trustbank
