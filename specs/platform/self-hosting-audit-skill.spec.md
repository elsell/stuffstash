# Self-Hosting Audit Skill Spec

## Purpose

Stuff Stash needs a repeatable Codex skill for auditing the public self-hosting experience from the perspective of a realistic technically capable homeowner.

The audit must find documentation gaps, hidden assumptions, setup friction, and reliability issues before real users encounter them.

## Scope

This spec covers the repo-local Codex skill at `.codex/skills/stuffstash-self-host-audit`.

It does not define production deployment architecture, hosting support policy, or application behavior. Those remain in the relevant product, platform, identity, media, and local-development specs.

## Persona

The skill must guide the agent to act as a homeowner with a family who is comfortable with command-line tools, Docker, and light self-hosting, but is not a Stuff Stash contributor and does not know the repository internals.

The persona should be skeptical, practical, and intolerant of avoidable friction.

## Rules

- The agent must ask the user for an SSH target before starting.
- The agent must ask the user for authentication details when they are needed and not already available through local SSH configuration.
- All setup, execution, hosting, browser testing, cloning, downloads, and verification must happen on the SSH target.
- The agent must get explicit user approval before using `sudo`, installing packages, changing firewall or network settings, starting long-lived services, or making other host-level changes that are not clearly part of an already-approved disposable audit environment.
- The agent must clean up or clearly report all containers, volumes, cloned repositories, scratch files, installed packages, and exposed ports left by the audit.
- The agent must not inspect Stuff Stash source code while executing the audit.
- The agent may use only the public GitHub Pages documentation site, the public README, release artifacts, and commands linked from those public sources.
- The canonical public documentation site is `https://elsell.github.io/stuffstash/`.
- The audit must exercise the documented Docker Compose self-hosting happy path
  with Postgres metadata persistence, datastore-backed SpiceDB authorization,
  Garage media storage, the static web container, and bundled Dex OIDC sign-in.
- The audit must not require a SQLite Docker Compose path. SQLite may remain an
  API runtime mode, but it is not a public self-hosting happy path unless a
  future spec creates a dedicated topology for it.
- The audit must include Dex OIDC authentication rather than accepting a
  development-only unauthenticated path.
- The audit must check whether the public docs make replacement of local Dex
  users, static clients, fixture passwords, and local HTTP origins clear before a
  household relies on the deployment.
- The audit must use a real browser automation flow, such as Playwright, for the user-facing web application.
- The browser flow must sign in, create or select a tenant, create or select an inventory, create items, and upload at least one image.
- Test images may come from public placeholder image services or generated local placeholder files when public services are unavailable.
- The audit must treat missing commands, missing prerequisites, unclear environment assumptions, flaky startup behavior, unsafe defaults, and deployment-only surprises as findings.
- The audit must redact secrets, tokens, cookies, authorization codes, private host details when requested, environment files, and sensitive authenticated screenshots or traces from chat output and saved artifacts.
- The audit must capture exact public inputs used for reproducibility, including documentation URLs, access timestamp, README revision or release version when known, container image references, and checksums for downloaded artifacts when available.
- The final report must be returned in chat.
- The report must not be committed. If the agent writes notes or logs, they must be placed in a gitignored location or outside the repository.

## Reporting Standard

The skill must tell the agent to be ruthless. It should not give documentation or setup flows the benefit of the doubt when a normal self-hoster would stall, guess, or silently accept an insecure configuration.

Findings must distinguish:

- blocking failures,
- high-friction or unreliable steps,
- security or production-readiness concerns,
- missing or misleading documentation,
- successful steps that are worth preserving.

Each finding should include the user-visible symptom, the exact step where it happened, the expected documentation or product behavior, and a concrete fix.
