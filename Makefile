.PHONY: test run compose-up compose-down docker-build docs-install docs-dev

GOCACHE ?= $(CURDIR)/.cache/go-build
PNPM ?= pnpm

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
	$(PNPM) --dir docs install --frozen-lockfile

docs-dev:
	$(PNPM) --dir docs dev
