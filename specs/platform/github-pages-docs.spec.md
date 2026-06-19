# GitHub Pages Documentation Deployment Spec

## Purpose

Stuff Stash needs the documentation site to publish automatically through GitHub Actions and GitHub Pages.

The deployment must also give maintainers a working documentation preview for pull requests and remove that preview when the pull request is done.

## Scope

This spec covers the Astro and Starlight documentation site under `docs/`, the GitHub Actions workflow that builds it, and the GitHub Pages publication layout.

It does not define application deployment, API hosting, or generated API client publication.

## Decisions

- The documentation site is built from `docs/` with Astro and Starlight.
- The canonical production site is `https://elsell.github.io/stuffstash/`.
- The production Pages content is published at the root of the `gh-pages` branch.
- Pull request previews are published under `pr-<number>/` on the same GitHub Pages site.
- Pull request preview URLs therefore use `https://elsell.github.io/stuffstash/pr-<number>/`.
- Pull request previews are deployed only for pull requests whose source branch is in the same repository, because forked pull requests do not receive a write-capable `GITHUB_TOKEN` and must not run trusted deployment code from untrusted changes.
- When a pull request is closed, the corresponding `pr-<number>/` directory must be removed from `gh-pages`.
- The workflow must use GitHub Actions and GitHub Pages only; no external preview hosting service is allowed.
- The workflow must not use the official `actions/deploy-pages` preview input while GitHub documents it as unavailable to the public.
- The workflow may publish by pushing the generated static files to the Pages branch.

## Astro Configuration

- `docs/astro.config.mjs` must read the public site origin from `STUFF_STASH_DOCS_SITE`.
- `docs/astro.config.mjs` must read the deployment base path from `STUFF_STASH_DOCS_BASE`.
- If these variables are not set, local builds must default to the default GitHub Pages origin and project base path.
- The base path must include leading and trailing slashes, such as `/stuffstash/` or `/stuffstash/pr-123/`.
- Production builds must use `/stuffstash/` as the base path.
- Pull request preview builds must use `/stuffstash/pr-<number>/` as the base path so Starlight links, scripts, styles, sitemap links, and asset URLs resolve under the preview directory.

## Workflow Requirements

- The workflow must run on pushes to `main`.
- The workflow must run on pull request open, synchronize, and reopen events.
- The workflow must run cleanup when pull requests close.
- The workflow must install dependencies with `pnpm install --frozen-lockfile`.
- The workflow must disable Astro telemetry.
- The workflow must fail if the docs build fails.
- The workflow must copy the built `docs/dist` output into the correct Pages branch directory.
- The workflow must preserve unrelated pull request preview directories during production deployments.
- The workflow must preserve production root content during pull request preview deployments.
- The workflow must remove only the matching pull request preview directory during cleanup.
- The workflow must create a `.nojekyll` file in the Pages branch so GitHub Pages serves Astro assets as generated.

## Pinned Tooling

- GitHub Actions dependencies must be pinned to immutable commit SHAs.
- Node must be pinned to a concrete version in the workflow.
- pnpm must use the `packageManager` version declared by `docs/package.json`.

## Verification

- Local verification must build the docs with the production project base path.
- Local verification must build the docs with a representative pull request preview base path.
- The generated preview HTML must reference CSS and script assets under the preview base path.
- After deployment, the published production URL and a pull request preview URL should be checked with `curl` for expected HTML and CSS asset availability.
