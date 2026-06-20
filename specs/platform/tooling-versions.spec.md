# Tooling Versions Spec

## Purpose

Stuff Stash pins build, runtime, and API tooling to reduce supply-chain risk and make generated artifacts reproducible.

## Scope

This spec tracks the first tooling versions used by the secure tracer bullet.

## Pinned Documentation Tooling

- Node.js for documentation CI: `24.17.0`.
- pnpm for documentation CI: `11.0.7`.
- Astro: `astro 6.4.4`.
- Starlight: `@astrojs/starlight 0.39.3`.
- Geist Sans documentation font package: `@fontsource/geist-sans 5.2.5`.

## Pinned Web And Client Tooling

- pnpm for web/client workspaces: `11.0.7`.
- SvelteKit: `@sveltejs/kit 2.63.0`.
- Svelte: `svelte 5.56.2`.
- Svelte Vite plugin: `@sveltejs/vite-plugin-svelte 7.1.2`.
- Svelte static adapter: `@sveltejs/adapter-static 3.0.10`.
- Vite: `vite 8.0.16`.
- TypeScript: `typescript 5.9.3`.
- Svelte check: `svelte-check 4.6.0`.
- Vitest: `vitest 4.1.8`.
- jsdom: `jsdom 29.1.1`.
- nanoid transitive override for Vite/PostCSS tooling: `nanoid 3.3.12`.
- Undici test-environment override for jsdom: `undici 7.27.1`.
- OpenAPI TypeScript generator: `openapi-typescript 7.13.0`.
- OpenAPI fetch runtime: `openapi-fetch 0.17.0`.
- shadcn-svelte CLI: `shadcn-svelte 1.3.0`.
- Tailwind CSS: `tailwindcss 4.3.0`.
- Tailwind Vite plugin: `@tailwindcss/vite 4.3.0`.
- Tailwind animation utilities: `tw-animate-css 1.4.0`.
- Bits UI: `bits-ui 2.18.1`.
- Class name helper: `clsx 2.1.1`.
- Tailwind class merge helper: `tailwind-merge 3.6.0`.
- Class variance helper: `class-variance-authority 0.7.1`.
- Tailwind variants helper: `tailwind-variants 3.2.2`.
- Inter variable font package for web UI: `@fontsource-variable/inter 5.2.8`.
- Internationalized date helper used by generated UI components: `@internationalized/date 3.12.2`.
- Svelte icon package used by generated UI components: `@lucide/svelte 1.17.0`.

## Pinned Go Dependencies

- Go module version: `go 1.25.8`.
- Huma: `github.com/danielgtaylor/huma/v2 v2.38.0`.
- ULID: `github.com/oklog/ulid/v2 v2.1.1`.
- Authzed Go client: `github.com/authzed/authzed-go v1.10.0`.
- OIDC verifier: `github.com/coreos/go-oidc/v3 v3.18.0`.
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
- Dex local OIDC service: `dexidp/dex:v2.44.0@sha256:5d0656fce7d453c0e3b2706abf40c0d0ce5b371fb0b73b3cf714d05f35fa5f86`.
- Garage local verification service, arm64: `dxflrs/garage:v2.3.0@sha256:2d3f94a89a8a02dc49fa75594d6df67ed9c6ffe08fe55ed023d0c9776f71a9bd`.
- Garage local verification service, amd64: `dxflrs/garage:v2.3.0@sha256:dac0c92add4f1a0b41035e94b41036a270ffbe88a37c7ac9c3f19e6dc5bdccf2`.

## Requirements

- Versions must not float.
- GitHub Actions must be pinned to immutable commit SHAs in workflow files.
- Container image overrides must still be pinned with `@sha256:`.
- New tools must be added here before use.
- Tooling changes must be atomic conventional commits.
- Generated artifacts must include drift checks once generation is introduced.
- Release version planning must be done by repository-owned scripts using Conventional Commit messages and SemVer. Do not add external release automation dependencies for version calculation.
- A release commit on `main` must only create a new tag and GitHub release when commits since the last SemVer tag require a release.
- Release image publication must produce immutable image digests. GitOps deployment state must prefer digests over mutable tags.
- Release builds must publish build provenance attestations for the container image before deployment automation consumes the image.
- Dependency freshness must be checked mechanically for npm and Go modules.
- npm package versions and Go module versions must be at least fourteen days old before they are accepted into the committed dependency graph.
- Dependency age checks must fail closed when package metadata cannot be retrieved or parsed, except for Go pseudo versions where the timestamp embedded in the version is available.
- The dependency age threshold may only be lowered or bypassed by a spec update that names the package, version, reason, and compensating verification.

## Pinned GitHub Actions

- `actions/checkout` `v6.0.1`: `8e8c483db84b4bee98b60c0593521ed34d9990e8`.
- `actions/setup-node` `v6`: `48b55a011bda9f5d6aeb4c2d9c7362e8dae4041e`.
- `actions/setup-go` `v6.1.0`: `4dc6199c7b1a012772edbd06daecab0f50c9053c`.
- `docker/setup-buildx-action` `v3.11.1`: `e468171a9de216ec08956ac3ada2f0791b6bd435`.
- `actions/attest-build-provenance` `v3`: `43d14bc2b83dec42d39ecae14e916627a18bb661`.
- `sigstore/cosign-installer` `v4.1.0`: `ba7bc0a3fef59531c69a25acd34668d6d3fe6f22`.
- `actions/upload-artifact` `v4.6.2`: `ea165f8d65b6e75b540449e92b4886f43607fa02`.
- `actions/download-artifact` `v5.0.0`: `634f93cb2916e3fdff6788551b99b062d0335ce0`.
