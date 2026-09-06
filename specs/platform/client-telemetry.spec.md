# Client Performance Telemetry

## Purpose

Measure mobile and web request and visible-image loading latency separately from
server work, without shipping Grafana credentials to clients.

## Contract and trust boundary

`POST /client-telemetry` accepts an authenticated batch of 1–50 measurements.
It is an operational, resource-free endpoint: it neither selects nor modifies a
tenant, inventory, asset, or access relationship. All authenticated principals may
report; there is no anonymous ingestion. Existing HTTP body and rate limits apply.
Do not create domain audit records for telemetry ingestion (which would feed back
into the operation being observed). Authentication still uses the normal port.

Each measurement contains only these bounded fields:

- `platform`: `ios`, `android`, `web`.
- `operation`: `request`, `image`.
- `surface`: `home`, `list`, `detail`, `gallery`, `fullscreen`, `upload`.
- `variant`: `none`, `small`, `medium`, `large`, `original`.
- `outcome`: `success`, `failure`, `cancelled`.
- `durationMs`: finite number from 0 through 60000.

Reject the entire batch if any measurement is invalid. No arbitrary fields,
identifiers, URLs, filenames, tokens, messages, or client-supplied event names are
exported. The authenticated principal is not attached to telemetry. Measurements
are untrusted client reports, not authoritative security or billing records.
Return the normal success envelope with the accepted count.

## Ports and adapters

A focused application package validates batches and records typed operational
events through the existing Observer port; the App facade only delegates. The
HTTP DTO and mapper remain separate from routes. The OTLP adapter maps validated
performance events to a seconds histogram and safe structured logs, with only
platform, operation, surface, variant and outcome as metric dimensions. Backend
request traces correlate ingestion logs; client durations remain distinct metrics.

Clients use injected performance observers and clocks. Their adapters send bounded
best-effort batches through the generated authenticated API client. Maximum pending
capacity is 100, batch size 20, flush interval 5s. No persistent/offline queue;
drop on transport failure and clear on session or server change. Never observe the
telemetry delivery request itself. A full buffer drops oldest events. Observers
must never delay or fail product requests, image display, or session shutdown.
Runtime configuration can disable client export; collection is enabled alongside
API observability for the measurement deployment. No backend credentials appear
in frontend configuration or bundles.

Record request duration at client adapter boundaries and image duration from
load-start to load/error in mounted visible surfaces. Do not count a component
unmount as success. Attribute cancellations distinctly and bound durations. Use
separate ios/android/web dimensions; browser tests are not physical-device proof.

## Verification

Test batch validation, bounded privacy fields, authenticated acceptance, anonymous
and malformed-token denial, and the absence of tenant/resource mutation. Use real
HTTP boundaries and fake observers. Test frontend batching, overflow, failure
isolation, session cleanup, and visible-image outcomes before wiring them. Generate
the OpenAPI client contract in CI and retain the ordinary security regression suite.
