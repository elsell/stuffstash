---
title: Security
description: The trust model for self-hosted household inventory.
---

Stuff Stash can hold private household data: photos, receipts, documents,
medicine, serial numbers, and shared access for other people. The security model
is part of the product, not a later hardening pass.

## Native SSO

Stuff Stash uses OIDC for authentication. The local stack uses Dex so the same
OIDC path can be tested without a real external provider.

The API verifies bearer ID tokens and fails closed when issuer, audience,
signature, expiry, or token shape is wrong. Provider-specific claims are
normalized at the adapter edge before they reach application behavior.

Google is the first planned external provider profile, and the architecture is
designed for standards-compliant OIDC issuers.

## Relationship-Based Access

Stuff Stash uses SpiceDB-style relationship authorization. Tenants are the
top-level boundary. Inventories live inside tenants and can be shared with
viewer or editor access.

Authorization checks happen at application boundaries. The same rules apply
whether a request comes from the web app, mobile app, REST API, a future agent
flow, or an import job.

Conversational inventory does not get special power. A model-assisted action can
only do what the signed-in user is allowed to do.

## Supply-Chain Discipline

The project keeps dependency and build inputs tight:

- Go, Node, pnpm, Astro, SvelteKit, Expo, OpenAPI tooling, and container images
  are pinned.
- Container base images are pinned by immutable digest.
- GitHub Actions are pinned to commit SHAs.
- npm installs ignore lifecycle scripts by default.
- npm and Go dependency updates must pass a minimum package-age check.
- Release images are built with SBOM and provenance metadata.
- Release images are signed and their signatures are verified in the pipeline.
- Build provenance attestations are published for release images.

## Local Secrets Are Local Only

Compose fixtures such as Dex users, local Postgres credentials, and local
SpiceDB `serve-testing` configuration exist for development and evaluation.
They are not production secrets and should not be reused in a deployed system.

## Data Portability

Security also includes exit. Stuff Stash is designed for JSON and CSV import and
export behind project-owned ports, so file formats and migration paths do not
leak into domain logic. Export must preserve tenant and inventory authorization
boundaries.
