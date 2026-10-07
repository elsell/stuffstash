---
title: Contributing
description: How to work on Stuff Stash without losing the product thread.
---

Stuff Stash is spec-driven. Specs are how the project keeps a fast-moving build
from drifting away from the product it is trying to become.

Before code changes, update the relevant spec in `specs/`. Code follows the
spec, not the other way around.

## Start With The Spec

Specs live in the top-level `specs/` directory and end in `.spec.md`.

Common areas include:

- `specs/assets/`
- `specs/locations/`
- `specs/identity-access/`
- `specs/agent-model/`
- `specs/platform/`

If a spec and code disagree, fix the spec first, then update the code.

## Keep Docs Selective

The public docs are not a mirror of every spec. Specs hold detailed product and
engineering decisions. Docs explain what a reader needs to understand, run,
self-host, trust, or contribute to Stuff Stash.

## Testing

Use test-driven development. Write real tests first, then implement the smallest
correct behavior, then refactor.

Tests should check behavior through the right boundary. Use fakes instead of
mocks. Security-sensitive behavior needs adversarial end-to-end tests at the real
interaction point.

## Local Checks

Run the main checks from the repository root:

```sh
make test
make web-test
make web-check
make docs-build
lefthook run pre-commit --all-files
```

Use narrower checks when you are working in one area, but run the relevant full
checks before opening a change.

## Commit Shape

Use atomic Conventional Commits. A commit should contain one coherent change:
the spec, tests, code, docs, and configuration needed for that change.

## Security-Sensitive Changes

Authentication, authorization, tenant isolation, sharing, imports, exports,
media, and conversational actions are security-sensitive. Changes in those areas
need adversarial tests for valid access, missing auth, wrong role, cross-tenant
access, bad tokens, expired tokens, and privilege-escalation attempts where they
apply.

Do not bypass ports or adapters to make a test pass. That is usually the bug the
architecture is trying to prevent.


## Client Copy And Localization

Put client messages in `packages/localization/src`, using a key that names the
screen or task. Components call their client's `t` adapter. Keep whole sentences
in the catalog and pass names or counts as named values. Do not translate IDs,
protocol values, imported content, or provider responses.

Use plural messages with `one` and `other` forms instead of appending “s”. The
shared formatter uses Intl plural rules and formats numbers for the device locale.
English is the shipping catalog; missing languages fall back to English.

For browser expansion checks, start the web app with
`VITE_STUFF_STASH_UI_LOCALE=en-XA`; use `ar-XB` for the RTL pseudolocale.
Native verification builds use `EXPO_PUBLIC_STUFF_STASH_UI_LOCALE`. These are
verification settings, not additional shipping translations. Reload after changing
the locale. Check that controls remain reachable and user-entered names stay intact.

The migration is still in progress. `node scripts/inventory-client-messages.mjs`
lists remaining literal candidates for review; it also finds technical errors and
constants, so its counts are not a list of confirmed UI defects.

## Publish CLI releases

The normal Release workflow attaches five portable CLI archives, individual
SHA-256 files, and `stuffstash-cli-release.json` to the project tag. The archive
includes third-party notices; `stuffstash version` reports its tag and source
commit. `version --json` reports whether that build includes a USB printer adapter;
this is separate from whether a printer is connected or ready. USB printer support
remains limited to Linux and the supported Brother
profile, even when ordinary inventory commands run on another platform.

If publication stops after staging its assets, run **Release → Run workflow**
on **main**, with `repair_run_id` set to the original Release workflow run ID.
Repair uses its retained `release-publication` artifact and exact tag/commit.
Draft-creation failures include the GitHub error status and validation details.
A lost response may still leave a draft behind; repair finds that draft and resumes
from the original assets.
It uploads missing assets and verifies existing bytes; it never overwrites a
mismatch. Investigate a mismatch or an expired artifact rather than rebuilding
an old tag with new source or dependencies.

Verified stable publication opens a maintenance PR for download links and
self-host image digests. The docs deploy after that PR merges. If its checks or
merge require attention, fix the PR; then run **Docs Pages** on **main** to refresh
the site. A draft, failed publication, or older repaired release cannot replace
newer stable download instructions.
