# GitHub Pages Documentation Deployment Spec

## Purpose

Stuff Stash needs the documentation site to publish automatically through GitHub Actions and GitHub Pages.

The deployment must also give maintainers a working documentation preview for pull requests and remove that preview when the pull request is done.

## Scope

This spec covers the Astro and Starlight documentation site under `docs/`, the GitHub Actions workflow that builds it, and the GitHub Pages publication layout.

It does not define application deployment, API hosting, or generated API client publication.

## Decisions

- The documentation site is built from `docs/` with Astro and Starlight.
- The canonical production site follows the repository Pages configuration, currently `https://stuffstash.org/`.
- The production Pages content is published at the root of the `gh-pages` branch.
- Pull request previews are published under `pr-<number>/` on the same GitHub Pages site.
- Pull request preview URLs append `pr-<number>/` to the configured production URL.
- Pull request previews are deployed only for pull requests whose source branch is in the same repository, because forked pull requests do not receive a write-capable `GITHUB_TOKEN` and must not run trusted deployment code from untrusted changes.
- When a pull request is closed, the corresponding `pr-<number>/` directory must be removed from `gh-pages`.
- The workflow must use GitHub Actions and GitHub Pages only; no external preview hosting service is allowed.
- The workflow must not use the official `actions/deploy-pages` preview input while GitHub documents it as unavailable to the public.
- The workflow may publish by pushing the generated static files to the Pages branch.

## Astro Configuration

- `docs/astro.config.mjs` must read the public site origin from `STUFF_STASH_DOCS_SITE`.
- `docs/astro.config.mjs` must read the deployment base path from `STUFF_STASH_DOCS_BASE`.
- If these variables are not set, local builds retain the default GitHub Pages origin and project base path; production always uses the repository Pages configuration.
- The base path must include leading and trailing slashes, such as `/stuffstash/` or `/stuffstash/pr-123/`.
- Builds read the repository Pages API using read-only Pages permission and derive
  the site origin and base from `html_url`. A custom domain uses `/`; without a
  custom domain, preserve the configured GitHub project path such as `/stuffstash/`.
  Invalid or unavailable configuration fails the build rather than deploying an
  assumed path. Preview builds append `pr-<number>/` to that base; summaries use
  the same derived public URL.
- Production publication restores the API-configured `CNAME` after replacing root
  content. Preview publication and cleanup preserve the root domain configuration.
- Critical checks cover custom-domain and project-path production/previews,
  generated CSS/script/image URLs, semantic tables, and responsive browser output.

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

## Planned Generated Printing Catalog

[Generated printer and label documentation](printing-catalog-docs.spec.md) requires
a Printing section generated from executable printer/template/media registries,
including PNG examples from the production renderer. PR generation and drift
checks, preview builds, and production publishing must include these outputs
without exposing deployment credentials to untrusted PR code. These outputs are implemented.

## Concurrent Pages publication

Production, preview publication, and preview cleanup share the Pages branch but
retain their existing per-preview workflow concurrency groups. Do not globally
serialize these workflows: GitHub concurrency can replace pending runs.

After creating its scoped content commit, each publisher attempts a normal push.
A non-fast-forward rejection fetches the latest Pages tip and rebases the local
commit before retrying, with at most five push attempts. The same helper is used
for deployment and cleanup. It never force-pushes or resolves conflicts by choosing
one side. A genuine content conflict fails visibly and aborts the rebase; other
push failures fail immediately with Git's diagnostics. Repeated contention fails
after the bounded attempts so the workflow can be rerun deliberately.

Disjoint preview/root updates must survive a retry. Preview publication must not
remove root files such as CNAME. Cleanup removes only its own preview. An identical
concurrent update may make the local commit redundant and still counts as success.
The first Pages-branch publication must also preserve a concurrently initialized
branch. Verify these cases using real temporary clones and a local bare Git remote,
including conflicting same-path writes and a rejecting receive hook. Tests must
not fake Git commands or use a network service.

The local Git publication fixtures disable automatic garbage collection and
maintenance for all Git commands, including publisher subprocesses and hooks.
Their repositories are short-lived; detached pack writers must not outlive the
fixture and race temporary-directory cleanup. This does not change production
Git settings or suppress cleanup failures.
