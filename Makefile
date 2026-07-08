.PHONY: test api-release-build go-structural-check run run-oidc-local run-spicedb spicedb-up spicedb-down dex-local-config dex-local verify-local-api verify-dex-oidc-api verify-mobile-oidc-pkce verify-mobile-oidc-pkce-local verify-spicedb-adapter verify-garage-blobstore verify-postgres-adapter verify-migrations migrate-up migrate-status compose-up compose-up-spicedb compose-up-oidc compose-up-oidc-lan compose-down docker-build docker-build-web dependency-age-check required-checks release-plan-test selfhost-happy-path-check scripts-test docs-install docs-dev docs-build docs-preview web-install web-dev web-build web-check web-test web-e2e-install web-e2e-test web-shadcn-check mobile-test mobile-check api-openapi-generate api-client-generate api-client-check api-client-test api-client-check-generated

GOCACHE ?= $(CURDIR)/.cache/go-build
STUFF_STASH_DATABASE_DSN ?= postgres://stuffstash:stuffstash-local@localhost:5432/stuffstash?sslmode=disable
STUFF_STASH_TEST_POSTGRES_DSN ?= postgres://stuffstash:stuffstash-local@localhost:5432/stuffstash?sslmode=disable
STUFF_STASH_LOCAL_HOST ?= localhost
STUFF_STASH_LOCAL_DEX_PORT ?= 5556
STUFF_STASH_LOCAL_API_PORT ?= 8080
STUFF_STASH_LOCAL_WEB_PORT ?= 5173
STUFF_STASH_LOCAL_DEX_CONFIG ?= .stuffstash/local/dex/config.yaml
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

go-structural-check:
	@base="$${STUFF_STASH_STRUCTURAL_BASE:-}"; \
	if [ -n "$$base" ] && git cat-file -e "$$base^{commit}" 2>/dev/null; then \
		go_files="$$(git diff --name-only --diff-filter=ACMRTUXB "$$base"...HEAD -- '*.go')"; \
	else \
		go_files="$$( { git diff --name-only --diff-filter=ACMRTUXB -- '*.go'; git diff --cached --name-only --diff-filter=ACMRTUXB -- '*.go'; git ls-files --others --exclude-standard -- '*.go'; } | sort -u)"; \
	fi; \
	if [ -z "$$go_files" ]; then \
		exit 0; \
	fi; \
	unformatted="$$(printf '%s\n' "$$go_files" | xargs gofmt -l)"; \
	if [ -n "$$unformatted" ]; then \
		printf '%s\n' "$$unformatted" >&2; \
		echo "Go files must be gofmt-formatted" >&2; \
		exit 1; \
	fi; \
	printf '%s\n' "$$go_files" | xargs scripts/check-go-structural-rules.sh; \
	printf '%s\n' "$$go_files" | xargs scripts/check-go-domain-imports.sh

run:
	GOCACHE=$(GOCACHE) go run ./apps/api/cmd/stuff-stash

run-oidc-local:
	STUFF_STASH_HTTP_ADDR=:$(STUFF_STASH_LOCAL_API_PORT) \
	STUFF_STASH_AUTH_MODE=oidc \
	STUFF_STASH_AUTHZ_MODE=memory \
	STUFF_STASH_REPOSITORY_MODE=memory \
	STUFF_STASH_CORS_ALLOWED_ORIGINS=http://$(STUFF_STASH_LOCAL_HOST):$(STUFF_STASH_LOCAL_WEB_PORT) \
	STUFF_STASH_OIDC_ISSUER=http://$(STUFF_STASH_LOCAL_HOST):$(STUFF_STASH_LOCAL_DEX_PORT)/dex \
	STUFF_STASH_OIDC_CLIENT_ID=stuff-stash-local \
	STUFF_STASH_OIDC_CLIENT_IDS=stuff-stash-local,stuff-stash-web-local,stuff-stash-mobile-local \
	STUFF_STASH_OIDC_MOBILE_CLIENT_ID=stuff-stash-mobile-local \
	STUFF_STASH_OIDC_MOBILE_REDIRECT_URI=stuffstash://auth/callback \
	STUFF_STASH_OIDC_MOBILE_SCOPES=openid,email,profile,offline_access \
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

verify-mobile-oidc-pkce:
	PATH="$(DOCS_PATH)" node scripts/verify-mobile-oidc-pkce.mjs

verify-mobile-oidc-pkce-local:
	STUFF_STASH_VERIFY_MOBILE_OIDC_ISSUER=http://$(STUFF_STASH_LOCAL_HOST):$(STUFF_STASH_LOCAL_DEX_PORT)/dex \
	STUFF_STASH_VERIFY_MOBILE_OIDC_API_BASE_URL=http://$(STUFF_STASH_LOCAL_HOST):$(STUFF_STASH_LOCAL_API_PORT) \
	PATH="$(DOCS_PATH)" node scripts/verify-mobile-oidc-pkce.mjs

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

dex-local-config:
	STUFF_STASH_WEB_ORIGIN=http://$(STUFF_STASH_LOCAL_HOST):$(STUFF_STASH_LOCAL_WEB_PORT) \
	STUFF_STASH_API_ORIGIN=http://$(STUFF_STASH_LOCAL_HOST):$(STUFF_STASH_LOCAL_API_PORT) \
	STUFF_STASH_DEX_ISSUER=http://$(STUFF_STASH_LOCAL_HOST):$(STUFF_STASH_LOCAL_DEX_PORT)/dex \
	STUFF_STASH_DEX_HTTP_ADDR=0.0.0.0:$(STUFF_STASH_LOCAL_DEX_PORT) \
	STUFF_STASH_DEX_CONFIG_OUT=$(STUFF_STASH_LOCAL_DEX_CONFIG) \
	STUFF_STASH_OIDC_MOBILE_REDIRECT_URI=stuffstash://auth/callback \
	PATH="$(DOCS_PATH)" node scripts/render-local-dex-config.mjs

dex-local: dex-local-config
	dex serve $(STUFF_STASH_LOCAL_DEX_CONFIG)

compose-up-oidc:
	STUFF_STASH_AUTH_MODE=oidc \
	STUFF_STASH_AUTHZ_MODE=spicedb \
	STUFF_STASH_CORS_ALLOWED_ORIGINS=http://localhost:5173 \
	STUFF_STASH_SPICEDB_TLS_ENABLED=false \
	STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA=true \
	STUFF_STASH_SPICEDB_SCHEMA_PATH=/deploy/spicedb/schema.zed \
	STUFF_STASH_OIDC_ISSUER=http://dex:5556/dex \
	STUFF_STASH_OIDC_CLIENT_ID=stuff-stash-local \
	STUFF_STASH_OIDC_CLIENT_IDS=stuff-stash-local,stuff-stash-web-local,stuff-stash-mobile-local \
	STUFF_STASH_OIDC_MOBILE_CLIENT_ID=stuff-stash-mobile-local \
	STUFF_STASH_OIDC_MOBILE_REDIRECT_URI=stuffstash://auth/callback \
	STUFF_STASH_OIDC_MOBILE_SCOPES=openid,email,profile,offline_access \
	docker compose -f compose.yaml -f compose.oidc.yaml up --build

compose-up-oidc-lan:
	@if [ -z "$(STUFF_STASH_LAN_HOST)" ]; then echo "Set STUFF_STASH_LAN_HOST to this machine's LAN IP, for example: make compose-up-oidc-lan STUFF_STASH_LAN_HOST=192.168.1.50" >&2; exit 1; fi
	STUFF_STASH_WEB_ORIGIN=http://$(STUFF_STASH_LAN_HOST):5173 \
	STUFF_STASH_API_ORIGIN=http://$(STUFF_STASH_LAN_HOST):8080 \
	STUFF_STASH_DEX_ISSUER=http://$(STUFF_STASH_LAN_HOST):5556/dex \
	STUFF_STASH_OIDC_MOBILE_REDIRECT_URI=stuffstash://auth/callback \
	PATH="$(DOCS_PATH)" node scripts/render-local-dex-config.mjs
	DEX_CONFIG_PATH=.stuffstash/local/dex/config.yaml \
	STUFF_STASH_AUTH_MODE=oidc \
	STUFF_STASH_AUTHZ_MODE=spicedb \
	STUFF_STASH_CORS_ALLOWED_ORIGINS=http://$(STUFF_STASH_LAN_HOST):5173 \
	STUFF_STASH_SPICEDB_TLS_ENABLED=false \
	STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA=true \
	STUFF_STASH_SPICEDB_SCHEMA_PATH=/deploy/spicedb/schema.zed \
	STUFF_STASH_OIDC_ISSUER=http://$(STUFF_STASH_LAN_HOST):5556/dex \
	STUFF_STASH_OIDC_CLIENT_ID=stuff-stash-local \
	STUFF_STASH_OIDC_CLIENT_IDS=stuff-stash-local,stuff-stash-web-local,stuff-stash-mobile-local \
	STUFF_STASH_OIDC_MOBILE_CLIENT_ID=stuff-stash-mobile-local \
	STUFF_STASH_OIDC_MOBILE_REDIRECT_URI=stuffstash://auth/callback \
	STUFF_STASH_OIDC_MOBILE_SCOPES=openid,email,profile,offline_access \
	docker compose -f compose.yaml -f compose.oidc.yaml up --build

compose-down:
	docker compose down

docker-build:
	docker build -t stuff-stash:local .

docker-build-web:
	docker build -f Dockerfile.web -t stuff-stash-web:local .

dependency-age-check:
	python3 scripts/check-dependency-age.py

required-checks: dependency-age-check scripts-test go-structural-check test api-release-build web-install web-test web-check web-build mobile-test mobile-check api-client-test api-client-check api-client-check-generated docs-install docs-build

release-plan-test:
	scripts/test-release-planner.sh

selfhost-happy-path-check:
	scripts/check-selfhost-happy-path.sh

scripts-test: release-plan-test selfhost-happy-path-check
	python3 -c 'import ast, pathlib; ast.parse(pathlib.Path("scripts/check-dependency-age.py").read_text(encoding="utf-8"))'
	python3 scripts/test-dependency-age.py
	PATH="$(DOCS_PATH)" node --check scripts/render-local-dex-config.mjs
	PATH="$(DOCS_PATH)" node --check scripts/update-selfhost-image-refs.mjs
	PATH="$(DOCS_PATH)" node --check scripts/verify-mobile-oidc-pkce.mjs

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

web-e2e-install:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web install:e2e-browsers

web-e2e-test:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web test:e2e

web-shadcn-check:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/web check:shadcn

mobile-test:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/mobile test

mobile-check:
	PATH="$(DOCS_PATH)" $(PNPM) --dir apps/mobile check

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
