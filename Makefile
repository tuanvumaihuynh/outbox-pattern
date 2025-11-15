ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

#########################
# Run
#########################
.PHONY: run
run:
	go run cmd/op-standalone/main.go


.PHONY: run-watch
run-watch:
	go run github.com/air-verse/air@v1.61.7 -c .air.toml

########################
# Docker
########################
.PHONY: compose-build
compose-build:
	docker compose build

.PHONY: compose-up
compose-up:
	docker compose up -d

########################
# Code generation
########################
.PHONY: gen-openapi
gen-openapi:
	set -eux

	pnpm --package=@redocly/cli@2.11.1 dlx redocly bundle ./api-contract/openapi.yml --output api-contract/openapi.gen.yml --ext yml
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1 \
		-config internal/http/gen/oapi-codegen.yml \
		api-contract/openapi.gen.yml

.PHONY: gen-sqlc
gen-sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate --file internal/storage/db/sqlc/sqlc.yml

.PHONY: gen-all
gen-all: gen-openapi gen-sqlc

#########################
# Database
#########################
GOOSE_DRIVER=postgres
GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
GOOSE_MIGRATION_DIR=internal/storage/db/migration

.PHONY: migrate-up
migrate-up:
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 up

.PHONY: migrate-down
migrate-down:
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 down

.PHONY: migrate-status
migrate-status:
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 status

.PHONY: migrate-create
migrate-create:
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 create "$(name)" sql

.PHONY: migrate-reset
migrate-reset:
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 reset

#########################
# Testing
#########################
.PHONY: test
test:
	go test -v --failfast ./...

.PHONY: test-cov
test-cov:
	go test -coverprofile=bin/coverage.out ./...
	go tool cover -html=bin/coverage.out -o bin/coverage.html
	@echo "Coverage report saved to bin/coverage.html"

########################
# Lint
########################
.PHONY: lint-go
lint-go:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.1 run ./... --config .golangci.yml
