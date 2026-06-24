---
title: Architecture
description: How Stuff Stash keeps product behavior separate from infrastructure.
---

Stuff Stash uses hexagonal architecture, also called ports and adapters. The
point is practical: the same inventory rules should work from the web app,
mobile app, REST API, conversational flow, future MCP tools, imports, and
background jobs.

Domain behavior sits in the center. Frameworks, databases, auth providers, model
providers, HTTP details, and blob storage stay outside the domain behind ports.

## The Core Rule

Moving an item should follow the same rules no matter how the command arrived.

That matters because Stuff Stash is voice-forward. Model output must never
bypass authorization, tenancy checks, validation, audit history, or normal
application services.

## Main Pieces

| Piece | Responsibility |
| --- | --- |
| Domain | Inventory concepts, rules, typed values, lifecycle behavior |
| Application services | Commands, queries, authorization checks, audit writes |
| Ports | Project-owned interfaces for persistence, auth, search, media, models, clocks, and observability |
| Adapters | Postgres/GORM, SpiceDB, OIDC, HTTP, blob storage, web/mobile clients |
| Specs | Source of truth for product and engineering decisions |

## Tenants And Inventories

A tenant is the top-level security boundary. An inventory lives inside one
tenant. Users may have access to one or more inventories, with relationships
such as owner, editor, or viewer.

This lets a self-hosted household support separate inventories for different
homes, collections, family members, or shared spaces without relying on loose
application roles.

## Assets, Containers, And Locations

Items, containers, and locations share one containment model.

That means a garage can contain a shelf, a shelf can contain a bin, and a bin can
contain an item. A toolbox can also contain items. The user-facing language stays
simple while the model stays flexible.

## Conversational Inventory

Conversational inventory is an interaction layer, not the domain core.

Speech-to-text, language models, and text-to-speech sit behind ports. A
conversation can search, ask clarifying questions, propose an action plan, and
execute approved application commands. It cannot write directly to persistence
or grant itself extra access.

## Why The Split Matters

The API and web app are separate deployables. Auth, authorization, persistence,
media, and model providers are swappable adapters. Generated OpenAPI types stay
at client adapter boundaries instead of becoming the frontend domain model.

That gives Stuff Stash room to support web, mobile, voice, imports, and future
agent workflows without each path inventing its own version of inventory logic.
