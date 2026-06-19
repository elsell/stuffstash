.PHONY: test run run-spicedb spicedb-up spicedb-down verify-local-api verify-spicedb-adapter verify-garage-blobstore verify-postgres-adapter verify-migrations migrate-up migrate-status compose-up compose-up-spicedb compose-down docker-build docs-install docs-dev docs-build docs-preview

GOCACHE ?= $(CURDIR)/.cache/go-build
STUFF_STASH_DATABASE_DSN ?= postgres://stuffstash:stuffstash-local@localhost:5432/stuffstash?sslmode=disable
STUFF_STASH_TEST_POSTGRES_DSN ?= postgres://stuffstash:stuffstash-local@localhost:5432/stuffstash?sslmode=disable
SPICEDB_CONTAINER ?= stuff-stash-spicedb
SPICEDB_GRPC_PORT ?= 50051
SPICEDB_PRESHARED_KEY ?=
SPICEDB_IMAGE ?= authzed/spicedb:v1.47.1@sha256:25c5499a43fdb206b7b1b72da4ba7ca911d92fd80d4d08ce2e95bf7ea0709788
CODEX_RUNTIME_BIN ?= $(HOME)/.cache/codex-runtimes/codex-primary-runtime/dependencies/bin
CODEX_RUNTIME_NODE_BIN ?= $(HOME)/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin
DOCS_PATH := $(CODEX_RUNTIME_NODE_BIN):$(PATH)
PNPM ?= $(shell command -v pnpm 2>/dev/null || test -x "$(CODEX_RUNTIME_BIN)/pnpm" && printf '%s\n' "$(CODEX_RUNTIME_BIN)/pnpm" || printf '%s\n' pnpm)

test:
	GOCACHE=$(GOCACHE) go test ./apps/api/...

run:
	GOCACHE=$(GOCACHE) go run ./apps/api/cmd/stuff-stash

run-spicedb: spicedb-up
	STUFF_STASH_AUTH_MODE=local-dev \
	STUFF_STASH_AUTHZ_MODE=spicedb \
	STUFF_STASH_SPICEDB_ENDPOINT=localhost:$(SPICEDB_GRPC_PORT) \
	STUFF_STASH_SPICEDB_PRESHARED_KEY=$(SPICEDB_PRESHARED_KEY) \
	STUFF_STASH_SPICEDB_TLS_ENABLED=false \
	STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA=true \
	STUFF_STASH_SPICEDB_SCHEMA_PATH=deploy/spicedb/schema.zed \
	GOCACHE=$(GOCACHE) go run ./apps/api/cmd/stuff-stash

spicedb-up:
	docker rm -f $(SPICEDB_CONTAINER) >/dev/null 2>&1 || true
	docker run --rm -d \
		--name $(SPICEDB_CONTAINER) \
		-p $(SPICEDB_GRPC_PORT):50051 \
		$(SPICEDB_IMAGE) \
		serve-testing

spicedb-down:
	docker rm -f $(SPICEDB_CONTAINER) >/dev/null 2>&1 || true

verify-local-api:
	scripts/verify-local-api.sh

verify-spicedb-adapter:
	scripts/verify-spicedb-adapter.sh

verify-garage-blobstore:
	scripts/verify-garage-blobstore.sh

verify-postgres-adapter:
	STUFF_STASH_TEST_POSTGRES_DSN="$(STUFF_STASH_TEST_POSTGRES_DSN)" GOCACHE=$(GOCACHE) go test ./apps/api/internal/adapters/gormstore -run TestPostgresStore -count=1

verify-migrations:
	STUFF_STASH_TEST_POSTGRES_DSN="$(STUFF_STASH_TEST_POSTGRES_DSN)" GOCACHE=$(GOCACHE) go test ./apps/api/internal/adapters/dbmigrations -run TestPostgresRunnerAppliesAndReportsNoopMigrations -count=1

migrate-up:
	STUFF_STASH_DATABASE_DSN="$(STUFF_STASH_DATABASE_DSN)" GOCACHE=$(GOCACHE) go run ./apps/api/cmd/stuff-stash migrate up

migrate-status:
	STUFF_STASH_DATABASE_DSN="$(STUFF_STASH_DATABASE_DSN)" GOCACHE=$(GOCACHE) go run ./apps/api/cmd/stuff-stash migrate status

compose-up:
	docker compose up --build

compose-up-spicedb:
	STUFF_STASH_AUTH_MODE=local-dev \
	STUFF_STASH_AUTHZ_MODE=spicedb \
	STUFF_STASH_SPICEDB_TLS_ENABLED=false \
	STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA=true \
	STUFF_STASH_SPICEDB_SCHEMA_PATH=/deploy/spicedb/schema.zed \
	docker compose up --build

compose-down:
	docker compose down

docker-build:
	docker build -t stuff-stash:local .

docs-install:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs install --frozen-lockfile

docs-dev:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs dev

docs-build:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs build

docs-preview:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs preview
