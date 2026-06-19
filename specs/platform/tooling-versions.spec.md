# Tooling Versions Spec

## Purpose

Stuff Stash pins build, runtime, and API tooling to reduce supply-chain risk and make generated artifacts reproducible.

## Scope

This spec tracks the first tooling versions used by the secure tracer bullet.

## Pinned Documentation Tooling

- Node.js for documentation CI: `24.17.0`.
- pnpm for documentation CI: `11.0.7`.
- Astro: `astro 6.4.8`.
- Starlight: `@astrojs/starlight 0.40.0`.

## Pinned Go Dependencies

- Go module version: `go 1.25.8`.
- Huma: `github.com/danielgtaylor/huma/v2 v2.38.0`.
- ULID: `github.com/oklog/ulid/v2 v2.1.1`.
- Authzed Go client: `github.com/authzed/authzed-go v1.10.0`.
- OIDC verifier: `github.com/coreos/go-oidc/v3 v3.19.0`.
- OAuth2 support: `golang.org/x/oauth2 v0.36.0`.
- gRPC: `google.golang.org/grpc v1.80.0`.
- GORM: `gorm.io/gorm v1.31.1`.
- GORM Postgres driver: `gorm.io/driver/postgres v1.6.0`.
- GORM SQLite driver: `gorm.io/driver/sqlite v1.6.0`.
- pgx Postgres driver: `github.com/jackc/pgx/v5 v5.6.0`.
- golang-migrate: `github.com/golang-migrate/migrate/v4 v4.19.1`.
- MinIO Go S3-compatible client: `github.com/minio/minio-go/v7 v7.2.0`.

## Pinned Container Images

- Go builder image: `registry.access.redhat.com/hi/go:1.25.10-builder-1780418048@sha256:1a99d42f555db97455998945faf3c797c1f65ce1b92e4d9952a589446d114d6c`.
- API runtime image: `registry.access.redhat.com/hi/core-runtime:2.42-1781714135@sha256:82ab1238082f405e19e1cc6e4950549371b6742ba6b649ca356c058249162540`.
- Postgres local service: `postgres:18.1-alpine@sha256:aa6eb304ddb6dd26df23d05db4e5cb05af8951cda3e0dc57731b771e0ef4ab29`.
- SpiceDB local service: `authzed/spicedb:v1.47.1@sha256:25c5499a43fdb206b7b1b72da4ba7ca911d92fd80d4d08ce2e95bf7ea0709788`.
- Garage local verification service, arm64: `dxflrs/garage:v2.3.0@sha256:2d3f94a89a8a02dc49fa75594d6df67ed9c6ffe08fe55ed023d0c9776f71a9bd`.
- Garage local verification service, amd64: `dxflrs/garage:v2.3.0@sha256:dac0c92add4f1a0b41035e94b41036a270ffbe88a37c7ac9c3f19e6dc5bdccf2`.

## Requirements

- Versions must not float.
- GitHub Actions must be pinned to immutable commit SHAs in workflow files.
- Container image overrides must still be pinned with `@sha256:`.
- New tools must be added here before use.
- Tooling changes must be atomic conventional commits.
- Generated artifacts must include drift checks once generation is introduced.

## Pinned GitHub Actions

- `actions/checkout` `v6.0.1`: `8e8c483db84b4bee98b60c0593521ed34d9990e8`.
- `actions/setup-node` `v6`: `48b55a011bda9f5d6aeb4c2d9c7362e8dae4041e`.
- `actions/upload-artifact` `v4.6.2`: `ea165f8d65b6e75b540449e92b4886f43607fa02`.
- `actions/download-artifact` `v5.0.0`: `634f93cb2916e3fdff6788551b99b062d0335ce0`.
