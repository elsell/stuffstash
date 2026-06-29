.PHONY: test api-release-build run run-spicedb spicedb-up spicedb-down verify-local-api verify-dex-oidc-api verify-spicedb-adapter verify-garage-blobstore verify-postgres-adapter verify-migrations migrate-up migrate-status compose-up compose-up-spicedb compose-up-oidc compose-down docker-build docker-build-web dependency-age-check release-plan-test scripts-test docs-install docs-dev docs-build docs-preview web-install web-dev web-build web-check web-test web-shadcn-check api-openapi-generate api-client-generate api-client-check api-client-test api-client-check-generated

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
PNPM ?= $(shell if command -v pnpm >/dev/null 2>&1; then command -v pnpm; elif test -x "$(CODEX_RUNTIME_BIN)/pnpm"; then printf '%s\n' "$(CODEX_RUNTIME_BIN)/pnpm"; else printf '%s\n' pnpm; fi)

test:
	GOCACHE=$(GOCACHE) go test ./apps/api/...

api-release-build:
	cd apps/api && GOWORK=off GOCACHE=$(GOCACHE) CGO_ENABLED=0 GOOS=linux go build -o /tmp/stuff-stash-release-check ./cmd/stuff-stash

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

verify-dex-oidc-api:
	scripts/verify-dex-oidc-api.sh

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

compose-up-oidc:
	STUFF_STASH_AUTH_MODE=oidc \
	STUFF_STASH_AUTHZ_MODE=spicedb \
	STUFF_STASH_CORS_ALLOWED_ORIGINS=http://localhost:5173 \
	STUFF_STASH_SPICEDB_TLS_ENABLED=false \
	STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA=true \
	STUFF_STASH_SPICEDB_SCHEMA_PATH=/deploy/spicedb/schema.zed \
	STUFF_STASH_OIDC_ISSUER=http://dex:5556/dex \
	STUFF_STASH_OIDC_CLIENT_ID=stuff-stash-local \
	STUFF_STASH_OIDC_CLIENT_IDS=stuff-stash-local,stuff-stash-web-local \
	docker compose -f compose.yaml -f compose.oidc.yaml up --build

compose-down:
	docker compose down

docker-build:
	docker build -t stuff-stash:local .

docker-build-web:
	docker build -f Dockerfile.web -t stuff-stash-web:local .

dependency-age-check:
	python3 scripts/check-dependency-age.py

release-plan-test:
	scripts/test-release-planner.sh

scripts-test: release-plan-test
	python3 -c 'import ast, pathlib; ast.parse(pathlib.Path("scripts/check-dependency-age.py").read_text(encoding="utf-8"))'

docs-install:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs install --frozen-lockfile

docs-dev:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs dev

docs-build:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs build

docs-preview:
	PATH="$(DOCS_PATH)" $(PNPM) --dir docs preview

web-install:
	PATH="$(DOCS_PATH)" $(PNPM) install --frozen-lockfile

web-dev:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web dev

web-build:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web build

web-check:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web check:shadcn
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web check

web-test:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web test

web-shadcn-check:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web check:shadcn

api-openapi-generate:
	GOCACHE=$(GOCACHE) go run ./apps/api/cmd/stuff-stash-openapi > packages/api-client/openapi.json

api-client-generate: api-openapi-generate
	PATH="$(DOCS_PATH)" $(PNPM) --dir packages/api-client generate

api-client-check:
	PATH="$(DOCS_PATH)" $(PNPM) --dir packages/api-client check

api-client-test:
	PATH="$(DOCS_PATH)" $(PNPM) --dir packages/api-client test

api-client-check-generated:
	PATH="$(DOCS_PATH)" PNPM="$(PNPM)" scripts/check-api-client-generated.sh
