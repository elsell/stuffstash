---
title: Overview
description: What Stuff Stash is and what problem it solves.
---

Stuff Stash is a home inventory system.

It tracks items, places, containers, photos, and history across one or more inventories. A tenant is the top-level security boundary. A tenant can have many inventories, such as household, tools, or medicine.

The main goal is low-friction updates. A user should be able to say something like:

> Move my fertilizer from the garage shelf to the wire rack.

The system should understand the request, check permissions, ask for confirmation when needed, and apply the same domain rules used by the normal API.

## What Exists Now

The repository has a small Go API scaffold with:

- A health endpoint.
- Domain-oriented observability through ports.
- Docker and Compose files for local work.
- Specs that define the product and architecture direction.

Most product behavior is still being specified before implementation.

## Main Building Blocks

- **API:** Go backend service.
- **Web:** SvelteKit, planned.
- **Mobile:** React Native with Expo, planned for iOS and Android.
- **Docs:** Astro and Starlight.
- **Authorization:** SpiceDB, planned.
- **Authentication:** OIDC and SSO, starting with Google, planned.
- **Storage:** PostgreSQL in production, SQLite for local development where useful.

