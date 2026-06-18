---
title: Specs And Process
description: How work happens in this repository.
---

Stuff Stash is spec-driven.

Before code changes, update the relevant spec in `specs/`. Code follows the spec, not the other way around.

## Specs

Specs live in top-level `specs/`.

Each file ends with `.spec.md` and is grouped by domain or platform area. Examples:

- `specs/assets/`
- `specs/locations/`
- `specs/identity-access/`
- `specs/platform/`

If a spec and code disagree, update the spec first.

## Testing

This project uses test-driven development.

Tests should check real behavior. Use fakes instead of mocks. Security-sensitive behavior needs adversarial end-to-end tests at the real boundary.

## Commits

Use atomic Conventional Commits.

Each commit should contain one coherent change: the spec, code, tests, docs, and config needed for that change.

## Security

Security is a primary concern.

All dependencies, tools, base images, and generated artifact sources must be pinned to reviewed versions where the ecosystem supports it. Container images must be pinned by digest.

Authentication and authorization behavior must be tested for real. Tests must cover valid access, unauthenticated access, unauthorized access, cross-tenant access, wrong-role access, bad tokens, expired tokens, and privilege escalation attempts where they apply.

## Documentation

Docs should help a newcomer understand the project and run it locally.

They should be short, direct, and useful. Do not document code that explains itself.

