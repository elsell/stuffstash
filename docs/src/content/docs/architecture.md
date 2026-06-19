---
title: Architecture
description: The project shape and the rules that matter most.
---

Stuff Stash uses hexagonal architecture, also called ports and adapters.

That means domain behavior sits in the center. Frameworks, databases, auth providers, model providers, and HTTP details stay outside the domain behind interfaces.

## The Rule

The core domain should not care how a command arrived.

Moving an item should follow the same rules whether the request came from:

- REST.
- The web app.
- The mobile app.
- A voice command.
- An MCP tool.
- A future import job.

## Bounded Contexts

The first contexts are:

- Assets.
- Inventories.
- Locations.
- Identity and access.
- Agent and model.
- Audit and history.
- Search.
- Media.
- Data portability.
- Expiration, still under discussion.

## Tenants And Inventories

A tenant is the top-level security boundary.

An inventory lives inside one tenant. Users can belong to more than one tenant and can have access to one or more inventories inside a tenant.

Authorization is relationship-based and uses the same shape as SpiceDB. The model should feel similar to Google Drive sharing: owners, editors, viewers, and direct sharing.

The current API slice proves this boundary with local development auth, OIDC token verification, SpiceDB authorization wiring, tenant creation, inventory creation, inventory listing, and Huma-generated OpenAPI.

When the API creates data that needs a SpiceDB relationship, it saves the data and an authorization outbox event in the same database transaction. The API then tries to drain that outbox right away. If SpiceDB is down, the relationship write can be retried instead of being lost.

## Assets And Locations

Assets and locations are separate concepts.

They share containment behavior. A garage shelf can contain an asset. A toolbox can also contain assets. The system should support this without turning every place into an item.

## Conversational Inventory

Voice and text commands are part of the main product experience. They are not part of the domain core.

Speech-to-text, language models, and text-to-speech all sit behind ports. Model output must never bypass authorization, tenancy checks, validation, audit history, or domain services.
