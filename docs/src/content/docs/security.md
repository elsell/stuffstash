---
title: Trust And Security
description: Why Stuff Stash is shaped for private household data.
---

Home inventory can include sensitive details: medicine, receipts, serial
numbers, documents, photos, and who has access to shared spaces.

Stuff Stash treats that data like it matters.

## Your Sign-In System

Stuff Stash uses OIDC for SSO. The API verifies bearer ID tokens and fails closed
when issuer, audience, signature, expiry, or token shape is wrong.

The Docker Compose self-host stack includes Dex so sign-in uses the same OIDC
boundary from the first run. Google is the first planned external provider
profile, and the same adapter shape supports standards-compliant OIDC issuers.

## Scoped Household Access

Tenants are the top-level security boundary. Inventories live inside tenants and
can be shared with viewer or editor access.

Relationship-based authorization keeps access checks explicit. A person can help
maintain one inventory without seeing every other inventory in the household.

Conversational actions use the signed-in user's permissions. The model-assisted
path does not get elevated access.

## A Safer Build Chain

Self-hosted software still depends on the build chain that produced it. Stuff
Stash keeps that chain tight:

- Dependencies, tools, base images, and GitHub Actions are pinned.
- Container base images are pinned by immutable digest.
- npm installs ignore lifecycle scripts by default.
- npm and Go dependency updates must pass a minimum package-age check.
- Release images are built with SBOM and provenance metadata.
- Release images are signed and verified in the pipeline.
- Build provenance attestations are published for release images.

## First-Run Secrets Are Not Household Secrets

The Compose self-host stack includes first-run Dex users, local Postgres
passwords, Garage keys, and SpiceDB credentials so you can start the product
without wiring outside infrastructure first.

Do not reuse those values in a deployed system.

The self-host stack uses datastore-backed SpiceDB, Postgres, and Garage volumes.
That makes normal container restarts durable when volumes are kept, but it does
not make the example secrets safe.

Replace first-run secrets before relying on a deployment:

- Dex users and static clients.
- Postgres password.
- Garage/S3 access and secret keys.
- Provider credential encryption key.
- SpiceDB credentials and datastore configuration.

## Exit Is Part Of Trust

Stuff Stash is designed for JSON and CSV import/export behind project-owned
ports. Export must preserve tenant and inventory authorization boundaries.

Data portability is not a bonus feature. It is part of the trust story.
