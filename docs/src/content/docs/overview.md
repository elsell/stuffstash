---
title: Overview
description: What Stuff Stash is and what problem it solves.
---

Stuff Stash is a home inventory system.

It is meant to track items, places, containers, photos, files, and history across one or more inventories. Today, the API handles tenants, inventories, assets, custom asset types, custom fields, asset attachments, audit history, and direct inventory sharing and revocation. A tenant is the top-level security boundary. A tenant can have many inventories, such as household, tools, or medicine.

The main goal is low-friction updates. A user should be able to say something like:

> Move my fertilizer from the garage shelf to the wire rack.

The system should understand the request, check permissions, ask for confirmation when needed, and apply the same domain rules used by the normal API.

## What Exists Now

The repository has a small Go API scaffold with:

- A health endpoint.
- Local development auth and production-shaped OIDC auth, with Dex for local OIDC verification.
- In-memory authorization for local use and SpiceDB authorization wiring.
- Tenant creation, inventory creation/listing, and first asset create/list/update/move flow.
- Custom asset types, such as medicine or tools, with type-specific custom fields.
- Tenant and inventory custom field definitions with asset value validation.
- Asset attachment upload, listing, and download with local filesystem and Garage-compatible blob storage.
- Asset search across the inventories a user can view.
- Durable audit history for the first state-changing actions.
- Direct inventory sharing by known principal ID, with viewer and editor grants and revocation.
- Huma-generated OpenAPI at `/openapi.json`.
- Domain-oriented observability through ports.
- Docker and Compose files for local work, including Postgres, SpiceDB, and an optional Dex OIDC override.
- Specs that define the product and architecture direction.

Most product behavior is still being specified before implementation.

## Main Building Blocks

- **API:** Go backend service.
- **Web:** SvelteKit, planned.
- **Mobile:** React Native with Expo, planned for iOS and Android.
- **Docs:** Astro and Starlight.
- **Authorization:** SpiceDB, with an in-memory adapter for fast local runs.
- **Authentication:** OIDC and SSO, with Dex for local verification and Google planned as the first external provider.
- **Storage:** PostgreSQL for metadata, Garage-compatible blob storage for media, local filesystem blob storage for development, and SQLite where useful for local-only fakes.
