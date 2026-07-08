---
name: stuffstash-self-host-audit
description: Audit the public Stuff Stash self-hosting journey from the perspective of a technically capable homeowner. Use when Codex is asked to test, critique, or report on Stuff Stash public documentation, README, durable Docker Compose setup, bundled Dex OIDC sign-in, Postgres/SpiceDB/Garage persistence, browser onboarding, image upload, and setup friction on a user-provided SSH target without inspecting source code.
---

# Stuff Stash Self-Host Audit

## Overview

Use this skill to perform a ruthless outside-in audit of the public Stuff Stash self-hosting experience. Act like a homeowner with a family who is comfortable with Docker, SSH, and light self-hosting, but who is not a Stuff Stash contributor and will not read the source code to fill documentation gaps.

## First Response

Ask for the SSH target before doing any work. Also ask for authentication details if local SSH configuration does not already handle the connection.

Request only what is needed:

- SSH target, such as `user@host` or an SSH config alias.
- SSH key path, password, port, or bastion details only if required.
- Whether the host is disposable and whether the user approves host-level changes such as `sudo`, package installation, firewall changes, starting long-lived services, exposed ports, or Docker volume creation.

Do not start local setup. All setup, execution, hosting, cloning, downloads, dependency installation, browser automation, and verification must happen on the SSH target.

## Non-Negotiables

- Do not inspect Stuff Stash source code during the audit.
- Use only the public documentation site, the public README, release artifacts, and commands linked from those public sources.
- Treat `https://elsell.github.io/stuffstash/` as the canonical public documentation site.
- Use the README only as a public fallback or comparison source.
- If the docs instruct cloning the repository, clone it on the SSH target only and run documented commands there. Do not open, search, or reason from source files.
- Do not patch Stuff Stash, edit compose files beyond documented user configuration, or infer private repository knowledge.
- Do not accept a development-only unauthenticated path as success.
- Include Dex OIDC sign-in in the verification.
- Exercise the documented Docker Compose self-host path with Postgres metadata persistence, datastore-backed SpiceDB authorization, Garage media storage, the static web container, and bundled Dex OIDC sign-in.
- Do not require or invent a SQLite Compose path. Treat any public documentation that presents SQLite as the self-host happy path as a finding.
- Use browser automation, preferably Playwright running against the remote-hosted app, for the end-user flow.
- Return the final report in chat. Do not commit it.
- If notes, screenshots, logs, or Playwright artifacts are written, place them outside the repository or in a gitignored scratch location and mention where they are.
- Redact secrets, tokens, cookies, authorization codes, private host details when requested, environment files, and sensitive authenticated screenshots or traces from chat output and saved artifacts.
- Keep artifact paths private to the audit user on the remote host when possible.
- Get explicit user approval before using `sudo`, installing packages, changing firewall or network settings, starting long-lived services, or making host-level changes unless the user has already confirmed the host is disposable and approved those actions.

## Audit Workflow

### 1. Establish Remote Access

Connect to the SSH target and record the remote baseline:

- OS and version.
- CPU architecture.
- Available memory and disk.
- Shell.
- Existing Docker, Compose, Node, package manager, Git, curl, and browser automation support.
- Whether `sudo` is available.

If a required dependency is absent, follow the public docs exactly. If the docs do not explain how to install it or assume it exists, log that as friction before using a reasonable installation path to continue.

Before installing anything or changing the host, pause for user approval unless that approval was already granted in the first response.

### 2. Gather Public Instructions

From the remote machine, read the public docs and README as a self-hoster would:

- Open or fetch the public documentation site.
- Locate self-hosting, Docker Compose, Postgres, datastore-backed SpiceDB, Garage, OIDC/Dex, web app, media upload, backup, upgrade, and troubleshooting guidance.
- Use the README only if the docs are incomplete or explicitly direct the user there.

Capture every place where the public path is ambiguous, missing, stale, internally inconsistent, or requires guessing.

Record reproducibility inputs:

- Exact public documentation URLs used.
- Access timestamp.
- README URL and revision if visible.
- Release, tag, artifact version, or container image reference used.
- Downloaded artifact checksums when public sources provide them or when computing them is straightforward.

### 3. Prepare A Clean Workspace

Create a disposable remote working directory outside any existing repository checkout. Clearly reset state between first-run and restart-durability attempts.

If the documented flow requires cloning the repository:

- Clone on the SSH target.
- Run only documented commands.
- Inspect only files explicitly named by public docs or README for self-hosting use, such as a Compose file or `.env.example`.
- Do not run exploratory file-tree inspection, `rg`, source browsing, package-manifest inspection, migration inspection, test inspection, or app/internal package inspection.
- Treat every need to inspect a repository file that was not explicitly named by public docs as a documentation failure.

### 4. Durable Compose Self-Host Path

Follow the public instructions to run Stuff Stash with Docker Compose, bundled Dex, Postgres, datastore-backed SpiceDB, Garage, the API, and the static web container.

Verify:

- Required environment variables are documented and understandable.
- Secrets and callback URLs are explained.
- Dex OIDC is configured and reachable as the default self-host identity provider.
- Public docs explain how to replace first-run Dex users, static clients, fixture passwords, and local HTTP origins before a household relies on the deployment.
- API and web services start reliably.
- Health or readiness checks are documented or discoverable from public docs.
- Data persists after container restart.
- Logs are useful when startup fails.
- The setup looks self-hostable, not merely a contributor dev loop.

Record exact commands used, where docs supplied them, and where you had to invent or repair anything.

### 5. Restart Durability Path

Restart the documented self-host stack without deleting volumes.

Verify:

- The same Dex user can sign back in.
- The tenant, inventory, location/container structure, item, and uploaded media remain visible.
- Postgres service configuration is documented.
- Migrations or schema setup are documented.
- Connection settings, credentials, volume persistence, and startup ordering are clear.
- Data persists after container restart.
- The self-host path does not accidentally rely on a contributor-only Vite server, `serve-testing` SpiceDB, in-memory auth, or hidden local state.

Treat hidden coupling between contributor evaluation commands and the self-host path as a finding.

### 6. Browser User Journey

Use Playwright from the remote machine, connecting to the remote-hosted web app. If a headed browser is not available, use headless mode and save screenshots or traces in scratch space.

Complete the public user flow:

- Open the web app.
- Sign in through Dex OIDC.
- Create or select a tenant.
- Create or select an inventory.
- Create at least one item.
- Create enough location or container structure to test normal household organization if the UI supports it.
- Upload at least one image attachment.
- Confirm the item and image remain visible after reload and service restart.

Use a public placeholder image service or a generated local placeholder image on the remote host. If image upload is unavailable, blocked, or undocumented, report the exact point of failure.

### 7. Production Readiness Checks

Assess the public self-hosting story as a real operator:

- Are required ports, origins, callback URLs, and reverse proxy assumptions clear?
- Are secrets generated, stored, and rotated safely?
- Is TLS or proxy guidance present when browser auth requires it?
- Are volumes, backups, restores, and upgrades described?
- Are version pins, image tags, and update expectations clear?
- Are logs, troubleshooting steps, and expected startup times documented?
- Are failure modes actionable for someone who cannot read source code?

Do not demand enterprise features, but do flag anything that makes a family self-hosting setup fragile, insecure, or hard to recover.

### 8. Cleanup And Residual State

Before final reporting, stop or remove audit-created services unless the user explicitly asks to keep them running.

Report residual state:

- Running containers and exposed ports left behind.
- Docker volumes, databases, uploaded media, and cloned repositories left behind.
- Scratch directories, browser artifacts, logs, and screenshots.
- Packages installed or host configuration changed during the audit.
- Any cleanup that could not be completed safely.

## Reporting

Be ruthless. Do not give the benefit of the doubt when a normal self-hoster would stall, guess, copy an insecure value, or silently get a local-only deployment.

Return a chat report with:

- `Executive verdict`: whether the current public self-hosting path is ready, partially ready, or not ready.
- `Environment`: SSH target alias or host description, OS, architecture, Docker/Compose versions, and any constraints.
- `Public inputs`: documentation URLs, access timestamp, README revision or release version if known, image references, and artifact checksums when available.
- `What worked`: short list of confirmed successful steps.
- `Blocking failures`: failures that prevented a complete self-hosted flow.
- `High-friction issues`: unclear, missing, flaky, or guess-heavy steps.
- `Security and production-readiness concerns`: auth, secrets, TLS, persistence, backups, upgrades, images, and exposed ports.
- `Documentation fixes`: concrete edits or pages needed.
- `Reproduction notes`: commands and steps, summarized enough for maintainers to replay without dumping giant logs.
- `Artifacts`: scratch paths for screenshots, traces, logs, or notes if any were written.
- `Cleanup`: what was removed, what remains, and why.
- `Open questions`: decisions the docs or product need from maintainers.

For each finding include:

- Severity.
- User-visible symptom.
- Exact step where it happened.
- Expected behavior or documentation.
- Concrete recommended fix.

Do not write the final report into the repository unless the user explicitly asks and the path is gitignored.
