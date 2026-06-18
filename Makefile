.PHONY: test run compose-up compose-down docker-build

GOCACHE ?= $(CURDIR)/.cache/go-build

test:
	GOCACHE=$(GOCACHE) go test ./...

run:
	GOCACHE=$(GOCACHE) go run ./cmd/stuff-stash

compose-up:
	docker compose up --build

compose-down:
	docker compose down

docker-build:
	docker build -t stuff-stash:local .
