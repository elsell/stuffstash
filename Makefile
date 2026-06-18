.PHONY: test run compose-up compose-down docker-build docs-install docs-dev docs-build docs-preview

GOCACHE ?= $(CURDIR)/.cache/go-build
CODEX_RUNTIME_BIN ?= $(HOME)/.cache/codex-runtimes/codex-primary-runtime/dependencies/bin
CODEX_RUNTIME_NODE_BIN ?= $(HOME)/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin
DOCS_PATH := $(CODEX_RUNTIME_NODE_BIN):$(PATH)
PNPM ?= $(shell command -v pnpm 2>/dev/null || test -x "$(CODEX_RUNTIME_BIN)/pnpm" && printf '%s\n' "$(CODEX_RUNTIME_BIN)/pnpm" || printf '%s\n' pnpm)

test:
	GOCACHE=$(GOCACHE) go test ./apps/api/...

run:
	GOCACHE=$(GOCACHE) go run ./apps/api/cmd/stuff-stash

compose-up:
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
