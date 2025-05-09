.PHONY: build run test migrate-up

# Build Vars
BINARY_NAME=goup
BUILD_DIR=build


# Go
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build

# SERVICE VARS
SERVICE=api

# DB migrate
MIGRATE=goose
MIGRATE_DIR=migrations
DB_DSN=postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=$${DB_SSLMODE}

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/$(SERVICE)
	@echo "Build $(BINARY_NAME) complete"

run:
	@echo "Running the application..."
	@$(GO) run ./cmd/$(SERVICE)

test:
	@echo "Running tests..."
	@$(GOTEST) ./... -v
	@echo "Tests complete"

migrate-up:
	@echo "Running database migrations..."
	$(MIGRATE) postgres $(DB_DSN) up -dir $(MIGRATE_DIR)
	@echo "Database migrations complete"

migrate-down:
	@echo "Rolling back database migrations..."
	@$(MIGRATE) postgres $(DB_DSN) down -dir $(MIGRATE_DIR)
	@echo "Database migrations rollback complete"
