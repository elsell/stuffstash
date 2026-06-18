# Tooling Versions Spec

## Purpose

Stuff Stash pins build, runtime, and API tooling to reduce supply-chain risk and make generated artifacts reproducible.

## Scope

This spec tracks the first tooling versions used by the secure tracer bullet.

## Pinned Go Dependencies

- Go module version: `go 1.25.8`.
- Huma: `github.com/danielgtaylor/huma/v2 v2.38.0`.
- ULID: `github.com/oklog/ulid/v2 v2.1.1`.
- Authzed Go client: `github.com/authzed/authzed-go v1.10.0`.
- OIDC verifier: `github.com/coreos/go-oidc/v3 v3.19.0`.
- OAuth2 support: `golang.org/x/oauth2 v0.36.0`.
- gRPC: `google.golang.org/grpc v1.80.0`.
- golang-migrate: `github.com/golang-migrate/migrate/v4 v4.19.1` when migration execution is implemented.

## Pinned Container Images

- Postgres local service: `postgres:18.1-alpine@sha256:aa6eb304ddb6dd26df23d05db4e5cb05af8951cda3e0dc57731b771e0ef4ab29`.
- SpiceDB local service: `authzed/spicedb:v1.47.1@sha256:25c5499a43fdb206b7b1b72da4ba7ca911d92fd80d4d08ce2e95bf7ea0709788`.

## Requirements

- Versions must not float.
- New tools must be added here before use.
- Tooling changes must be atomic conventional commits.
- Generated artifacts must include drift checks once generation is introduced.
