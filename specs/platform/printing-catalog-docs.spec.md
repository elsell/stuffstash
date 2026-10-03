# Generated Printer And Label Documentation

## Status And Scope

Implementation candidate October 3, 2026. Offline executable registry export,
production-rendered examples, owned-output drift checks, and docs build wiring
are implemented in the candidate; physical printing remains unverified. The Astro/Starlight documentation site
must have a Printing section showing first-party printers, supported label sizes,
and templates with generated PNG examples. The initial supported hardware/media
scope is exactly Brother QL-800 over USB on Linux with 29 x 90 mm labels; no
additional printer or label size is required. The generator must handle future
registrations, but must not invent support entries to populate its pages. The
catalog must follow executable registrations, not a second manually maintained compatibility list.

See [printer integration](../printing/printer-integration.spec.md),
[asset labels](../printing/asset-labels.spec.md), and [CLI](cli.spec.md).
This does not change the first-release printer registration flow: each printer
is registered together with one manually selected size. Public size listings
are reference documentation, not a user-managed media catalog in the app.

## Source Of Truth

- Each built-in printer adapter registers typed public descriptors alongside its
  executable adapter registration: stable ID, display name/model, supported host
  platforms/transports, shipped media preset IDs/versions, status capabilities,
  setup guide reference, and implementation/verification status.
- Each media preset declares the physical label size, printable bounds, raster
  resolution/orientation/color mode, and constraints already used for validation
  and rendering. Documentation translates these into useful units and plain
  language; users are not asked to understand protocol fields.
- Each template registration declares ID/version, display name, short purpose,
  supported options/defaults, and executable compatibility/rendering behavior.
  Use those same registrations to populate runtime template selection and docs.
- Export descriptors from the actual API renderer/template registry and CLI
  built-in adapter registry through deterministic, offline commands. Exporting
  must not require a running server, database, credentials, USB device, or network.
  Do not scan Go source text to infer support or maintain a docs-only registry.
- Application bootstrap and exporter tests must prove the exported registration
  set equals the built-in runtime set. An unwired adapter file is not supported
  merely because its name appears in a descriptor.
- Keep generation at the tooling/application boundary. Domain packages must not
  import Astro/Markdown tooling, and the CLI must not import the API's internal
  packages. Separate registry exporters may produce versioned JSON combined by
  the docs generator; no deployment coupling or shared mutable global registry.
- First-party means shipped and maintained with Stuff Stash. It does not mean
  every listed combination has passed physical-device tests. Distinguish released,
  experimental/unverified, and deprecated support, with concise limitations and
  checked-in evidence references where available. Never upgrade a status solely
  because a PNG renders or the registry contains a model name.
- Initial QL-800/Linux/USB/29 x 90 mm support is planned in this spec, not evidence
  of a shipped adapter. Until implementation exists, public docs must not present
  it as working support. Generated PR previews describe the candidate branch;
  production docs follow the existing publishing policy and label unreleased
  capabilities clearly rather than implying an already downloadable release.

## Public Pages And PNG Examples

- Add a Printing sidebar section with a short human-written setup/workflow guide
  and generated catalog pages for Supported printers, Label sizes, and Templates.
  Explain how to select the correct size at registration and edit it after a roll
  change. Keep operational guidance curated; do not publish the internal specs.
- Planned generated pages live in
  `docs/src/content/docs/printing/supported-printers.mdx`, `label-sizes.mdx`, and
  `templates.mdx`. Generated PNGs live in `docs/src/assets/printing/generated/`.
  Link setup guidance from printer entries; keep model/platform/transport/size
  and known limitations easy to scan. Avoid one thin page per registry entry.
- Show one PNG for each supported built-in template/default-options and media
  preset combination. Deduplicate identical presets shared by multiple printers;
  do not render every impossible Cartesian combination. Include additional
  representative images only for named options that materially change layout,
  such as hiding the human-readable reference; exhaustive option permutations
  are unnecessary.
- Generate previews through the same production label renderer and compatibility
  validation used for print jobs. Use fixed synthetic content, a reserved example
  URL/identity, and reviewed pinned fonts. Do not use real inventory data, real
  credentials, private endpoints, AI images, screenshots of a mockup, or a
  separate simplified documentation renderer.
- Sample content must not imply that a template always fits. Explain that long
  names/URLs or smaller media can affect layout; unsupported combinations are
  excluded or shown with their actual compatibility reason. A failure rendering
  an advertised supported fixture fails generation rather than disappearing.
- Keep the printer-target raster artifact unmodified. For the human-facing PNG,
  apply only the renderer's declared display-orientation transform, so the user
  sees the label in reading orientation. No redraw, rescaling, QR interpolation,
  or omitted margins. This distinguishes a feed-oriented QL-800 raster from its
  horizontal on-label appearance without inventing a second layout engine.
- Each image includes nearby template/version, label dimensions, and option
  selection, meaningful alt text, intrinsic pixel dimensions, and a downloadable
  original. Use responsive display without claiming on-screen physical scale.
  Keep monochrome labels legible on both site themes. PNG QR contents must decode
  to the fixed fixture URL; examples do not identify real inventory assets.
- Preserve existing base-path handling for GitHub Pages and PR previews. Use
  Astro-managed asset URLs; no hard-coded production or root-only image paths.

## Reproducible Generation And Pull Requests

- Add `make printing-docs-generate` to export built-in registries, validate their
  cross-references/versions, render PNG fixtures, and write generated pages plus
  an owned-output manifest. Generation is deterministic and offline: stable
  ordering/names, pinned dependencies/fonts, fixed locale/timezone and fixtures,
  no clock-dependent data or nondeterministic PNG metadata.
- Check generated Markdown/MDX, PNGs, and the manifest into the repository so
  registry/rendering changes are reviewable with their documentation in a PR.
  Generated output carries a source/generation notice; never edit it by hand.
- Add `make printing-docs-check-generated` to generate into an isolated temporary
  tree and fail on changed, missing, extra, or stale owned output, including binary
  PNG differences. Remove obsolete files only within manifest-owned generated
  paths; never delete hand-written guides or unrelated site assets.
- PR CI runs generation/drift checking automatically, including when adapter or
  template registrations, presets, renderer code, fonts, fixtures, generator code,
  or build configuration change. Prefer an unconditional lightweight catalog
  check to incomplete path filters. The actual PNG render/check may use precise
  dependency triggers only if tests prove those triggers cannot miss changes.
- CI builds the docs from the newly generated candidate output and exposes the
  affected pages/images through the existing PR preview or build artifact path.
  A drift failure provides the regeneration command and generated output artifact
  so the contributor can commit the update. CI must not silently pass using stale
  checked-in catalog pages.
- No privileged bot push or automatic PR commit is required: automatic generation,
  rendered preview, and a failing drift check enforce synchronized changes.
  Run untrusted PR generation without deployment credentials; use existing
  permission-separated publishing for previews, especially fork PRs.
- Wire the same check into relevant pre-commit hooks and root required checks.
  Production documentation builds also generate/check the catalog so a registry
  change cannot publish an outdated support table. Do not mutate source files
  unexpectedly as part of the read-only check target.

## Required Verification

- Tests cover adding/removing/renaming an adapter or template, a preset/version
  change, a template layout change, and a font/fixture change. Verify that affected
  pages and PNGs update, removed entries disappear, and unrelated pages survive.
- Verify runtime/export registry parity, duplicate IDs, missing preset references,
  invalid dimensions, and unsupported declared capabilities fail clearly.
- Generate twice on the pinned supported build environment and require identical
  file names and bytes. CI drift tests must detect extra/missing images as well
  as stale text; rendering failure cannot be swallowed.
- Decode every generated PNG's QR and verify exact dimensions, orientation,
  non-overlapping content, and readable representative title/reference cases.
  Rendering tests and documentation examples do not establish physical printing.
- Build Astro/Starlight at production and PR-preview base paths; verify generated
  page navigation, asset URLs, accessible image text, mobile layout, light/dark
  themes, and download links. Visually inspect representative images and pages.
- During implementation, documentation setup instructions must be executed where
  possible, with unverified hardware behavior explicitly marked. Keep outstanding
  user-device checks in the shared user-testing checklist without blocking
  independent delivery or claiming those tests passed.

## CLI Download Link Integration

Generated printer pages link to the automatically updated CLI install section
specified in [CLI release binaries](cli.spec.md#github-release-binaries-and-download-documentation).
That section renders a pinned `curl` command for the newest successfully published
stable CLI release from verified release metadata, including checksum steps.
Catalog generation remains offline; release publication owns metadata refresh.
Do not advertise a future adapter as supported by an older downloadable binary.

## Delivery tooling

`stuffstash printers catalog --json` exports the same constructed adapter set used
by CLI runtime bootstrap. The built-in Brother adapter and API catalog map the
same project-owned `printingprofiles` media values; protocol print-head offsets
remain private to the adapter. The docs command deduplicates these exported media
snapshots and passes them to the existing offline `label-catalog` production
renderer exporter. No source-text scanning or copied physical presets are used.

The generator owns only its three catalog pages, manifest and generated binary
asset directory. A read-only check renders into a temporary tree, reports added,
changed, missing and obsolete owned paths, and can retain the candidate output
for PR artifacts. Rendering and registration errors fail the check. Curated setup
instructions and other docs are never removed by regeneration.

CLI release provenance includes the first-party profile module as a source input.
The archive preserves existing project licensing: include a root LICENSE if one
exists and identify the first-party source without inventing a license grant. The
license collector may recognize only that exact checked-in first-party module path,
not skip arbitrary replacement modules. A dirty profile prevents a release build.
