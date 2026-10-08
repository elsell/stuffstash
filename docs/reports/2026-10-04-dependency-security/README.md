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

## gRPC follow-up

GitHub subsequently reconciled the original baseline, leaving its three explicitly
patched upstream alerts plus this additional advisory:

| Advisory | Manifest | Reviewed resolution | Validation |
| --- | --- | --- | --- |
| [GHSA-2v4p-qf9q-27wj](https://github.com/grpc/grpc-go/security/advisories/GHSA-2v4p-qf9q-27wj) | `apps/api/go.mod` | gRPC 1.83.2, with upstream-required x/net 0.58.0, x/crypto 0.55.0 and x/text 0.41.0 | PR #411 required checks and connected authorization CI; alert closure requires merge |

The vulnerable path is an xDS gRPC server interceptor. Stuff Stash's adapter
constructs a SpiceDB client; no application xDS server entry point was found.
The patch removes the vulnerable library version without changing product or
permission behavior. The original JSON matrix remains the historical baseline.

## October 7: remaining formatter and patched dependencies

Alert #185 adds `sprintf-js` through React Native → babel-jest →
babel-plugin-istanbul → @istanbuljs/load-nyc-config → js-yaml 3 → argparse 1.
The upstream formatter's latest 1.1.3 remains affected by
[unbounded numeric precision](https://github.com/alexei/sprintf.js/issues/237).
The loader calls js-yaml `load`, which the already pinned 4.3.2 supports. A scoped
`@istanbuljs/load-nyc-config>js-yaml` override removes the legacy formatter chain;
the installed-library regression reads YAML coverage settings and executes actual
Babel-instrumented code. The regenerated lock contains no sprintf-js entry, and all five installed-library
security regressions pass, including the new YAML/instrumentation check. The formatter is development tooling here;
that exposure assessment does not dismiss the dependency alert.

The other three alerts retain their existing patches and installed-library
regressions. Registry and upstream source checks on October 7 found:

- Braces remains at 3.0.3 and micromatch at 4.0.8. Metro and Jest still use that
  chain. [PR #72](https://github.com/micromatch/braces/pull/72) is closed without
  merge; retain the bounded-nesting patch rather than remove its protection.
- Node-forge remains at 1.4.0. Expo code-signing-certificates 0.0.7 still depends
  on it; updating Expo tooling does not itself remove the RSA verification risk.
  [PR #1152](https://github.com/digitalbazaar/forge/pull/1152) remains open.
- Http-cache-semantics 4.3.0 exists, but its published source changes response
  status and Vary handling, not the max-stale reuse guard. It retains the same
  vulnerable branch addressed by [issue #56](https://github.com/kornelski/http-cache-semantics/issues/56).
  Do not replace patched 4.2.0 with unpatched 4.3.0 merely because 4.3.0 is outside
  the advisory's current version range. Astro's use remains build-time image
  cache lifetime calculation; retain the cache-revalidation patch.

These are explicit upstream blockers, not closed alerts. No alert is dismissed
and no package identity is changed to hide a finding.

### Patched releases in this batch

| Alerts | Dependency | Exact update |
| --- | --- | --- |
| #182, #184 | source-map-js | 1.2.2 in both workspaces |
| #183 | compression | 1.8.2 in the application workspace |
| #181 | postcss-selector-parser | 7.1.6 in the documentation graph and shared override |
| #185 | sprintf-js | Removed through the scoped YAML loader update above |

Web and documentation builds passed locally. Native exports and type checks were
interrupted; repository CI provides the remaining checks before merge. Independent
code review found no confirmed issues. GitHub closure must be checked after merge;
no remaining alert is dismissed.
