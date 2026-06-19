#!/usr/bin/env bash
set -euo pipefail

container="${SPICEDB_INTEGRATION_CONTAINER:-stuff-stash-spicedb-integration}"
network="${SPICEDB_INTEGRATION_NETWORK:-stuff-stash-spicedb-integration}"
port="${SPICEDB_INTEGRATION_PORT:-15051}"
image="${SPICEDB_IMAGE:-authzed/spicedb:v1.47.1@sha256:25c5499a43fdb206b7b1b72da4ba7ca911d92fd80d4d08ce2e95bf7ea0709788}"
go_test_image="${GO_BUILDER_IMAGE:-registry.access.redhat.com/hi/go:1.25.10-builder-1780418048@sha256:1a99d42f555db97455998945faf3c797c1f65ce1b92e4d9952a589446d114d6c}"

cleanup() {
  docker rm -f "$container" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
}

run_spicedb_integration_test() {
  local image="$1"

  if command -v go >/dev/null 2>&1; then
    STUFF_STASH_SPICEDB_INTEGRATION_ENDPOINT="localhost:${port}" \
      GOCACHE="${GOCACHE:-$PWD/.cache/go-build}" \
      go test ./apps/api/internal/adapters/spicedb -run TestSpiceDBIntegration -count=1
    return
  fi

  docker run --rm \
    --network "$network" \
    -e "STUFF_STASH_SPICEDB_INTEGRATION_ENDPOINT=${container}:50051" \
    -e GOCACHE=/tmp/go-build \
    -v "$PWD:/src" \
    -w /src \
    "$image" \
    go test ./apps/api/internal/adapters/spicedb -run TestSpiceDBIntegration -count=1
}

cleanup
trap cleanup EXIT

docker network create "$network" >/dev/null

docker run --rm -d \
  --name "$container" \
  --network "$network" \
  -p "${port}:50051" \
  "$image" \
  serve-testing >/dev/null

run_spicedb_integration_test "$go_test_image"
