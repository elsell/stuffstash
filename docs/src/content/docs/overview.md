---
title: Overview
description: What Stuff Stash is and what problem it solves.
---

Stuff Stash is a home inventory system.

It is meant to track items, places, containers, photos, files, and history across one or more inventories. Today, the API handles tenants, inventories, assets, custom asset types, custom fields, asset attachments, audit history, direct inventory sharing, invite-link tokens, pending-invitation management, and revocation. A tenant is the top-level security boundary. A tenant can have many inventories, such as household, tools, or medicine.

The main goal is low-friction updates. A user should be able to say something like:

> Move my fertilizer from the garage shelf to the wire rack.

The system should understand the request, check permissions, ask for confirmation when needed, and apply the same domain rules used by the normal API.

## What Exists Now

The repository has a Go API and the first separate SvelteKit web app with:

- A health endpoint.
- Local development auth and production-shaped OIDC auth, with Dex for local OIDC verification.
- In-memory authorization for local use and SpiceDB authorization wiring.
- Tenant and inventory creation, browsing, update, archive, restore, and hard-delete flows.
- Asset create/list/detail/update/move/archive/restore/delete flow.
- Custom asset types, such as medicine or tools, with type-specific custom fields.
- Tenant and inventory custom field definitions with asset value validation.
- Asset attachment upload, listing, and download with local filesystem and Garage-compatible blob storage.
- Asset search across the inventories a user can view.
- Durable audit history for state changes and selected reads.
- Inventory sharing by known principal ID or invite-link token, with viewer and editor access, pending-invitation management, and revocation.
- Huma-generated OpenAPI at `/openapi.json`.
- A generated TypeScript API client used by the web app at its adapter boundary.
- A web tracer bullet for local Dex sign-in, inventory creation, asset creation, and asset browsing.
- Domain-oriented observability through ports.
- Docker and Compose files for local work, including Postgres, SpiceDB, and an optional Dex OIDC override.
- Specs that define the product and architecture direction.

Most product behavior is still being specified before implementation.

## Main Building Blocks

- **API:** Go backend service.
- **Web:** SvelteKit.
- **Mobile:** React Native with Expo, planned for iOS and Android.
- **Docs:** Astro and Starlight.
- **Authorization:** SpiceDB, with an in-memory adapter for fast local runs.
- **Authentication:** OIDC and SSO, with Dex for local verification and Google planned as the first external provider.
- **Storage:** PostgreSQL for metadata, Garage-compatible blob storage for media, local filesystem blob storage for development, and SQLite where useful for local-only fakes.
