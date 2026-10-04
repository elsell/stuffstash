# Dependency Security Remediation — October 4, 2026

## Scope and baseline

The authenticated GitHub Dependabot inventory contains 151 open alerts: 2 critical,
80 high, 54 medium, and 15 low. These represent 96 distinct advisories across 35
package/ecosystem pairs. The affected graphs are the root pnpm workspace, separate
Astro documentation workspace, and API Go module. Direct-manifest alerts duplicate
some lockfile findings; every alert still needs an explicit disposition.

Upgrade affected dependencies to exact reviewed patched releases and regenerate
lockfiles with pinned tooling. Preserve mobile SDK/native alignment, authentication,
rendering, printing, and existing functionality. Do not dismiss alerts, disable
security checks, rename vulnerable packages, or insert incompatible overrides to
hide findings. Prefer compatible maintained parent upgrades. A necessary major
migration must preserve and validate its affected contract.

## Initial reviewed upgrade targets

- SvelteKit 2.70.2 and Vitest 4.1.11 retain their existing major lines.
- Astro 7.3.5, Starlight 0.42.2, and lucode-starlight 1.0.0 form the compatible
  documentation migration. Preserve custom-domain/project-path asset resolution,
  semantic MDX tables, generated printer previews, and CLI download pages.
- gRPC 1.83.1, OpenTelemetry stable modules 1.45.0 and log modules 0.21.0,
  and golang.org/x/net 0.57.0 address the Go advisory ranges. Keep API/CLI module
  graphs and release exporter builds consistent; validate required Go versions.
- Review and pin patched transitive versions for brace-expansion, xmldom,
  devalue, undici, image-size, js-yaml, nanoid, postcss, selector-parser,
  browserslist, baseline-browser-mapping, decode-uri-component, cookie, uuid,
  shell-quote, esbuild, sharp, smol-toml, and svgo. Preserve separate major lines
  when patched branches exist. Verify consumers when crossing a major boundary.

## Advisories without published fixes

At baseline, braces 3.0.3, http-cache-semantics 4.2.0, and node-forge 1.4.0
are their latest published releases and remain affected. Investigate maintained
parent versions or compatible replacements that safely eliminate those paths.
If no supported path exists, a reviewed narrow patch requires an exploit-oriented
regression and explicit provenance. Such a patch mitigates behavior but does not
make GitHub's affected-version alert disappear; report that remaining condition
without claiming closure. Do not remove product capabilities to remove a package.

## Verification and reporting

Maintain one per-alert matrix with manifest, advisory, package, affected range,
selected version or mitigation, and verification evidence. Inspect PR #362 rather
than blindly merging its partial dependency changes. Run affected unit/behavioral
checks, web and mobile type checks, both native Metro exports, Go API/CLI tests,
docs/custom-domain builds and generated artifact checks. Security regressions must
exercise actual libraries or faithful implementations, not mocks. Run repository
required checks and the code critic before proposing merge. GitHub alert closure
can only be verified after the default branch receives the remediation.

## Temporary upstream mitigations

Apply reviewed package patches while retaining the original affected package names
and versions: braces PR #72 at `28d440b5dd449dbf1fe6f3506cf94ecca4d02660`
limits nested parser/AST traversal; node-forge PR #1152 at
`ceba34402e329f0365134f23fe19898756527d65` rejects extra nested DigestAlgorithm
members. For http-cache-semantics issue #56, non-storable and no-cache responses,
shared cookie-bearing responses
without explicit public/immutable opt-in and proxy-revalidate responses must never
be reused through request max-stale. Verify malformed RSA structures, excessive
brace nesting, and shared-cache reuse against the actual installed libraries,
alongside legitimate signatures, ordinary brace expansion, and ordinary caching.
Keep these three alerts explicitly unresolved until upstream fixed versions can
replace the patches. No affected-version suppression is permitted.

OpenTelemetry log 0.21 uses the shared attribute value types. Migrate only the
telemetry adapter calls, preserving typed field allowlists and emitted event bodies.

The separate docs workspace explicitly includes its own root package and runs
installation from its directory so pinned pnpm resolves its reviewed overrides and security patches independently of the
application workspace. CI runs the installed-library security regressions after
installing both workspaces.

Metro 0.83.7 remains aligned with Expo 55. Its asset adapter must read image bytes
before calling image-size 2, whose API no longer accepts filesystem paths. Retain
Metro archive handling and validate actual PNG metadata and both native exports.

Astro 7.3.5 and its matched compiler 0.5.1/Markdown 0.4.2/MDX 8.0.2
packages, plus smol-toml 1.9.0, require explicit temporary age-policy exceptions
for security fixes. Retain older compatible Vite 8.3.0/Rolldown 1.2.9, Starlight
0.42.2, magic-string 1.4.1 and Undici 8.10.2 rather than widening those exceptions.

## gRPC advisory follow-up

After the original baseline reconciles, GHSA-2v4p-qf9q-27wj requires gRPC
1.83.2 in the API module. The advisory concerns xDS servers receiving requests
without authority/Host headers. Stuff Stash uses a gRPC client for SpiceDB and
has no application `xds.NewGRPCServer` entry point; still remove the affected
library version rather than suppress the alert. Accept gRPC's required reviewed
x/net 0.58.0, x/crypto 0.55.0 and x/text 0.41.0 graph updates (and x/tools 0.48.0
checksum resolution). Keep API/CLI workspace builds consistent and retain real
SpiceDB authorization and required-check coverage. No new application endpoint or
authentication behavior is introduced. Verify GitHub closure after merge.
