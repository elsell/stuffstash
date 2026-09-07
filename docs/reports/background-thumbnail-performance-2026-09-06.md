# Background thumbnail production comparison

Background generation is deployed. Keep **one worker** at the current API limit of
500m CPU and 512Mi memory. Two workers delivered batches faster but worsened ordinary
API latency, immediate-open latency, and memory usage. Concurrency remains configurable.

Images opened after a fixed preparation window now have a much faster median.
Freshly uploaded images remain a regression; this is not a claim that every image
loading path improved. Intermittent blob-storage stalls also remain unresolved.

## Final results

Times below are median / p95, using nearest-rank quantiles. Each completed run has
54 fixed-delay thumbnail reads, 18 warm reads, and six immediate small-thumbnail
reads. The reference is `0d84c2da9`; both worker settings use `f66f2d965`.

| Operation | Reference, on-demand | One worker, selected | Two workers |
| --- | ---: | ---: | ---: |
| Thumbnails after 30-second preparation window | 7.489 / 23.189 s | **0.329 / 20.491 s** | 0.334 / 11.784 s |
| Warm small thumbnails | 61 / 95 ms | 60 / 918 ms | 59 / 198 ms |
| Small thumbnail immediately after upload | 3.401 / 8.456 s | 8.116 / 22.723 s | 11.068 / 35.543 s |
| Ordinary authenticated asset list | 102 / 878 ms | 95 / 715 ms | 111 / 1,200 ms |
| Peak sampled API working set | 329 MiB | 335 MiB | 401 MiB |
| API restarts during completed run | 0 | 0 | 0 |

The selected one-worker run reduced fixed-delay median latency by **95.6%** and
p95 by **11.6%** relative to the reference. Its warm-read median was unchanged,
but the warm-read tail and immediate-open latency worsened. Ordinary-request counts
were 346, 392 and 303 respectively; those distributions include both active image
work and idle preparation windows, rather than a separate saturation test.

For each batch of 18 thumbnail responses, total delivery times were:

| Configuration | Batch 1 | Batch 2 | Batch 3 | Aggregate delivery rate |
| --- | ---: | ---: | ---: | ---: |
| Reference | 72.621 s | 78.861 s | 93.092 s | 0.221 thumbnails/s |
| One worker | 16.569 s | 31.143 s | 59.404 s | 0.504 thumbnails/s |
| Two workers | 17.712 s | 3.326 s | 24.221 s | 1.193 thumbnails/s |

Delivery rate is 54 responses divided by the sum of the three read-phase durations;
it is not isolated background-job throughput. Two workers improved this rate but
raised ordinary-request p95 by 68% and immediate-open p95 by 56% compared with one.
That fails the agreed condition for adopting two without hurting foreground requests.
Sampled CPU peaks were about 480m and 502m; both settings approached the same CPU cap.

Both final runs completed with no request failures. All **156** thumbnail response
hashes matched the reference outputs and had valid JPEG signatures. The first
notification-image one-worker attempt did fail on an upload, as detailed below;
the completed repeat must not be read as proof that storage failures are resolved.

## Method and practical limits

The authenticated experiment ran from `paul` against `api.stuffstash.jsksell.com`,
using the same six JPEG originals copied into dedicated experiment assets. Sizes
were 3,755,993; 2,859,202; 2,790,703; 2,170,153; 2,086,071; and 2,058,024 bytes.
Private evidence records original and response SHA-256 hashes. User originals were
unchanged; each run deleted its own experiment assets.

Each configuration received three batches of six sequential JSON/base64 uploads.
After the last upload, the driver waited 30 seconds, then requested small, medium
and large thumbnails with HTTP concurrency two. Warm small reads followed. A
separate six-image cohort requested small thumbnails immediately after upload.
An authenticated asset-list request ran roughly once per second throughout.

Both final worker runs used the same driver, image, telemetry, single deployment,
CPU/memory limits and storage. Existing-image backfill was drained first: all 277
jobs completed, with none pending, leased or failed. One worker ran before two;
shared production storage and time-of-run variation are not controlled. The
small fixed corpus and three batches do not establish broad statistical certainty.
Kubernetes working-set samples are not instantaneous allocation peaks.

Baseline uploads included external readiness probes. Those probes were removed
from final runs after timing out and adding storage traffic; they stopped before
baseline thumbnail reads. Consequently upload timing is diagnostic, not an isolated
before/after claim. Fixed-delay upload median/p95 was 2.969/10.165 s for the reference,
5.897/20.113 s with one worker and 6.267/16.505 s with two.

Final reads were correlated with existing API cache-source events. All 78 request
trace IDs matched in each run: one worker produced 57 cache / 21 generated events;
two produced 72 cache / 7 generated events, including one duplicated request trace.
Events are not assumed to be one per request. A cache event may follow waiting for
publication and does not prove the image was ready at request start.

This measures API-visible loading, not physical mobile UI rendering. Mobile uses
authenticated small/medium/large references; web attachment and primary-photo views
use small references. Both benefit through the unchanged API contract without a
frontend release. JSON uploads do not substitute for direct-upload timing tests;
automated boundary tests cover direct-upload completion and imports.

## Implementation and verification

Uploads and imports durably enqueue work with attachment creation. An in-process Go
worker claims the database queue, downloads/decodes once, and publishes small,
medium and large derivatives incrementally. Existing images use a resumable backfill.
The shared admission port limits foreground and background processing and gives
foreground waiters priority. Leases, bounded retries, guarded publication, permanent
storage-key reservations and deletion rechecks protect restart and deletion races.

Readers subscribe through a publication-notification port before their first cache
lookup. Successful publication wakes waiting readers for one lookup; they retain
one cancellable admission attempt for on-demand fallback. The in-process notification
adapter holds only active subscriptions and can be replaced behind the port.

An intermediate polling implementation was rejected after production traces showed
repeated 250 ms storage timeouts. Regression CI `34073300611` reproduced seven reads
while waiting instead of one. Backend race tests, real PostgreSQL coordination tests,
structural checks and API publication passed [CI 34073917277](https://github.com/elsell/stuffstash/actions/runs/34073917277).
Code critic review found no blocking issue. Local builds/hooks were not run because
of disk constraints; the relevant Go structural and test checks ran in CI.
The separate client-telemetry job still fails on the known mobile TypeScript
`node:module` resolution issue. No frontend telemetry release is included.

The deployed image is
`ghcr.io/elsell/stuffstash@sha256:9ea5c73cb635f826008dab4fc66f63b4d2ea50eb1a3161df2f9971b84ed7de7a`.
GitOps `12c5f93` deployed it with one worker, `729b365` activated two for comparison,
and `1156b0d` restores one. The single-replica deployment uses Recreate; no Deployment
was manually patched. A prematurely started two-worker experiment was stopped when
the old process still reported one, cleaned up and excluded. Configuration checksum
and explicit concurrency annotation now agree with the selected setting.

## Failures and remaining work

The first notification-image one-worker attempt stopped in batch two on upload
HTTP 502 after 30.571 s. Its trace recorded 27.292 s in original blob write and
30.552 s in the HTTP operation, exceeding the configured 30-second write timeout.
Its 24 completed thumbnail responses matched reference hashes. The owned asset was
removed before the unchanged-protocol repeat.

An earlier two-worker test during backfill also stopped on upload HTTP 502, with
30.452 s in blob write. A prior one-worker run during backfill completed but had
immediate-open p95 of 45.142 s. These loaded runs are not the final controlled
comparison because older jobs differed as backfill progressed. Temporary two-CPU
catch-up was excluded; 500m was restored before final measurements.

Garage uses `nfs-csi` storage for `/var/lib/garage`. That is an investigation lead,
not proof that NFS caused the stalls. Storage layout and HTTP timeout settings were
not changed. Immediate-open contention and blob-storage latency need follow-up;
raising concurrency at the current CPU limit is not supported by these results.

## Deferred observability

Web/mobile telemetry rollout, physical-device rendering measurements, broader
dependency instrumentation, alerts and SLO dashboards, and extended profiling are
deferred. Existing API traces, metrics, logs and profiling were sufficient to identify
the polling issue and verify this comparison.

Private reproducibility evidence is under the operator's measurement directory:
`background-baseline-0d84c2da9.jsonl`,
`background-notify-one-repeat-f66f2d965.jsonl`, and
`background-notify-two-confirmed-f66f2d965.jsonl`, with resource samples, cache events,
corpus hashes and the unchanged driver. Failed and excluded attempts are retained.
