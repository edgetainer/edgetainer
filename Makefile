# Makefile for Edgetainer

.PHONY: build-server build-agent build-all clean run-server run-agent \
	docker-build-server docker-build-agent docker-build-all \
	docker-run-server docker-run-agent docker-clean

# Go build flags
BUILD_FLAGS := -ldflags="-s -w"
GOBUILD := go build $(BUILD_FLAGS)

# Binary output directories
BIN_DIR := bin
SERVER_BIN := $(BIN_DIR)/edgetainer-server
AGENT_BIN := $(BIN_DIR)/edgetainer-agent

# Source directories
SERVER_SRC := cmd/server
AGENT_SRC := cmd/agent

# Default target
all: build-all

# Create directories
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# Build the server binary
build-server: $(BIN_DIR)
	$(GOBUILD) -o $(SERVER_BIN) ./$(SERVER_SRC)

# Build the agent binary
build-agent: $(BIN_DIR)
	$(GOBUILD) -o $(AGENT_BIN) ./$(AGENT_SRC)

# Build all binaries
build-all: build-server build-agent

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)

# Run the server
run-server: build-server
	$(SERVER_BIN) --config config/server-config.yaml

# Run the agent
run-agent: build-agent
	$(AGENT_BIN) --config config/agent-config.yaml

# Generate a development SSH key for testing
gen-ssh-key:
	ssh-keygen -t rsa -b 4096 -f dev_ssh_key -N ""

# Start a development PostgreSQL server using Docker
start-dev-db:
	docker run --name edgetainer-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=edgetainer -p 5432:5432 -d postgres:13

# Stop the development PostgreSQL server
stop-dev-db:
	docker stop edgetainer-postgres
	docker rm edgetainer-postgres

# Initialize the development environment
init-dev: gen-ssh-key start-dev-db

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Docker commands
docker-build-server:
	docker build -t ghcr.io/edgetainer/edgetainer/server:latest -f docker/Dockerfile.server .

docker-build-agent:
	docker build -t ghcr.io/edgetainer/edgetainer/agent:latest -f docker/Dockerfile.agent .

docker-build-all: docker-build-server docker-build-agent

docker-run-server:
	docker-compose -f compose.yml up -d

docker-run-agent:
	docker-compose -f compose.agent.yml up -d

docker-clean:
	docker-compose -f compose.yml down -v
	docker-compose -f compose.agent.yml down -v

# Setup server and agent in development mode
docker-setup: docker-build-all docker-run-server

# Show logs for the server
docker-logs-server:
	docker-compose -f compose.yml logs -f edgetainer-server

# Show logs for the agent
docker-logs-agent:
	docker-compose -f compose.agent.yml logs -f edgetainer-agent

# Get the agent's public key
docker-agent-pubkey:
	docker-compose -f compose.agent.yml exec edgetainer-agent cat /app/ssh/id_rsa.pub

# Register an agent with the server
register-agent:
	@if [ -z "$(id)" ]; then \
		echo "Error: Missing agent ID. Usage: make register-agent id=<agent-id>"; \
		exit 1; \
	fi
	./scripts/register_agent.sh $(id)
