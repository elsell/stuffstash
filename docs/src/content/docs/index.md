---
title: Stuff Stash
description: Home inventory that is easy enough to keep current.
template: splash
hero:
  tagline: Find the thing you put in a storage bin last year.
  actions:
    - text: See it in action (coming soon)
      link: product/
    - text: Read the docs
      link: self-hosting/
---

Stuff Stash is a self-hosted home inventory app built around the moment that
usually makes inventory software fail: you need one thing, you know you own it,
and you do not remember where you put it.

The goal is simple: open the app, start voice mode, and ask where the thing is.
The app should understand your household language, search what you are allowed
to see, and help you update the inventory without turning every move, refill,
or cleanup into a data-entry chore.

## Built For Real Homes

- Add items, containers, locations, and photos quickly.
- Ask where something is instead of remembering exact labels.
- Move and update items conversationally, with review before anything saves.
- Search manually when voice is not the fastest path.
- Sign in with SSO and share inventories with the right people.
- Import and export your data so the inventory is not trapped in one tool.

## Self-Hosted By Design

Stuff Stash separates the Go API from the SvelteKit web app. That makes the
deployment shape explicit: run the API, run the web app, and connect them to
Postgres, SpiceDB, and your OIDC provider.

Docker Compose is the easiest path. The same split also fits Kubernetes
deployments: run the API, run the web app, and wire each service through runtime
configuration.

## Trust Matters

Household inventory can include receipts, photos, medicine, documents, serial
numbers, and access for family members or guests. Stuff Stash treats that as a
security boundary, not an afterthought.

The project uses native OIDC/SSO, relationship-based authorization, pinned
dependencies and base images, dependency-age checks, signed container images,
and provenance attestations.
