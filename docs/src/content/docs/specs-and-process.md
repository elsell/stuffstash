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
