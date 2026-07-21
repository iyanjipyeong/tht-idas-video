include .env

MIGRATION_SRC_DIR ?= migrations
MIGRATION_DB_DRIVER := postgres
MIGRATION_DB_PORT ?= 5432
MIGRATION_DB_SSLMODE ?= disable
MIGRATION_DB_PG_SCHEMA ?= public
MIGRATION_URL := "$(MIGRATION_DB_DRIVER)://$(DATABASE_USERNAME):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(MIGRATION_DB_PORT)/$(DATABASE_NAME)?sslmode=$(MIGRATION_DB_SSLMODE)&search_path=$(MIGRATION_DB_PG_SCHEMA)"
CMD__MIGRATE := migrate -source "file://$(MIGRATION_SRC_DIR)" -database $(MIGRATION_URL)

.PHONY: serve build utest mocking clean-mocks migrate-up migrate-down migrate-script db-script db-version

GO_TEST_ENV := GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod
UTEST_FLAGS ?= -short -v
BUILD_ID ?= dev-local
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
APP_BUILD_DIR ?= bin
APP_BINARY ?= $(APP_BUILD_DIR)/idas-video-api
LDFLAGS := -X 'idas-video/internal/infrastructure/buildinfo.BuildID=$(BUILD_ID)' -X 'idas-video/internal/infrastructure/buildinfo.BuildTime=$(BUILD_TIME)'
MOCKERY_VERSION ?= v2.53.5
GO_RUN_ENV := GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod
SERVE_ENV := LOG_LEVEL=info LOG_NAMESPACE=idas-video $(GO_RUN_ENV)
MOCKERY ?= $(GO_RUN_ENV) go run github.com/vektra/mockery/v2@$(MOCKERY_VERSION)
MOCKERY_OUTBOUND_DIR ?= internal/usecase/outbound
MOCKERY_INTERFACE ?= IRepositoryContext
MOCKERY_OUTPUT_FILE ?= mock_i_repository_context.go

serve:
	$(SERVE_ENV) go run -ldflags="$(LDFLAGS)" ./cmd/api/main.go

build:
	@mkdir -p $(APP_BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(APP_BINARY) ./cmd/api/main.go

utest:
	$(GO_TEST_ENV) go test $(UTEST_FLAGS) ./...

mocking:
	@$(MOCKERY) --dir $(MOCKERY_OUTBOUND_DIR) --name $(MOCKERY_INTERFACE) --inpackage --case=underscore --filename $(MOCKERY_OUTPUT_FILE)

clean-mocks:
	@rm -f $(MOCKERY_OUTBOUND_DIR)/$(MOCKERY_OUTPUT_FILE)

migrate-up:
	$(CMD__MIGRATE) up

migrate-down:
	$(CMD__MIGRATE) down 1

migrate-script:
	@read -p "Enter migration script name: " name; \
		migrate create -ext sql -dir "$(MIGRATION_SRC_DIR)" -seq "$$name"

db-script: migrate-script

## db-version: Force set version to fix dirty state
db-version:
	@read -p "Enter migration version: " ver; \
		$(CMD__MIGRATE) force $$ver
