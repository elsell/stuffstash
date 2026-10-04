# Dependency security remediation

The October 4 inventory contains **151 open GitHub alerts**, covering 96 advisories
and 35 package/ecosystem pairs. The candidate graph removes affected versions for
148 alerts. Three have no published fixed version and remain open; reviewed package
patches mitigate them without disguising their names or versions. No alerts were
dismissed. [alerts.json](alerts.json) records every alert and selected version.
GitHub closure is unverified until this work reaches the default branch.

## Three upstream blockers

| Dependency | Upstream state and reachability | Temporary mitigation |
| --- | --- | --- |
| braces 3.0.3 | Latest micromatch 4.0.8 still includes it in Metro/build tooling. Untrusted nested glob input can exhaust recursion. | Runtime changes from [upstream PR 72](https://github.com/micromatch/braces/pull/72), commit `28d440b5dd449dbf1fe6f3506cf94ecca4d02660`, bound parser and AST traversal depth. |
| node-forge 1.4.0 | Current Expo CLI and code-signing certificates packages still include it for certificate/signature tooling. | Narrow RSA DigestAlgorithm member validation from [upstream PR 1152](https://github.com/digitalbazaar/forge/pull/1152), commit `ceba34402e329f0365134f23fe19898756527d65`. |
| http-cache-semantics 4.2.0 | Astro 7.3.5 still includes it for build-time remote-image TTL calculation. That path does not call the vulnerable `evaluateRequest` reuse API. | Local patch for [upstream issue 56](https://github.com/kornelski/http-cache-semantics/issues/56): no-cache/non-storable and restricted shared responses cannot bypass revalidation using max-stale or stale-while-revalidate. |

Upgrading to current maintained parents does not eliminate these dependencies.
Replacing Metro, Expo certificate tooling, or Astro to remove them would require
larger compatibility work. Keep the patches visible, monitor upstream releases,
and replace them with fixed versions when available. This is mitigation, not a
claim that GitHub's three version alerts are closed.

## Compatibility and evidence

- Actual installed-library regressions fail before and pass after the three
  mitigations. They preserve valid RSA signatures, ordinary brace expansion, and
  legitimate fresh/stale caching. A fourth regression exercises real PNG metadata
  through Metro; its small adapter patch supplies bytes to image-size 2.
- API and CLI Go suites pass. OpenTelemetry log values migrate to shared attribute
  types while retaining existing safe-field tests. Generated Go and TypeScript
  clients and generated printing pages/PNGs have no drift. Real SpiceDB adapter
  and HTTP authorization integration suites pass against an isolated pinned server.
- Client type checks pass. Web: 187 files / 1,235 tests. Mobile: 375 files / 2,192
  tests. One initial concurrent mobile run missed a rendered press target; its
  focused 44-test file and subsequent full suite passed. No test was disabled.
- Fresh production Hermes exports pass for Android and iOS after the Metro patch.
  This verifies packaging, not physical native app behavior or App Store upload.
- Astro 7 builds successfully with semantic generated tables. Chromium at 390 and
  1,280 pixels verifies table overflow containment, loaded PNG previews and CLI
  download links on both custom-domain and project-path configurations.
- Security regressions run in `make required-checks` and the relevant pre-commit
  hook. Pinned age exceptions are limited to the required Astro security toolchain
  and smol-toml fix; compatible older transitives remain pinned. The complete
  dependency-age check passes. Script and structural checks pass; API and CLI
  release builds pass. Local API build validation disables VCS stamping because
  an unrelated empty parent `.git` directory confuses Go worktree discovery.

The code critic reviewed the patches, compatibility changes and telemetry migration;
confirmed findings are resolved. Final CI, default-branch alert closure, and any
new release remain pending.
