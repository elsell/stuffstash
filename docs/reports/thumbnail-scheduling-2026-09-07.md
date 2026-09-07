# Thumbnail scheduling optimization

Cooperative scheduling reduced the median immediate open after back-to-back uploads
from **8.04 to 4.82 seconds (40%)** in the controlled three-photo experiment. It did
not eliminate long waits: the slowest of six back-to-back opens increased from
12.50 to 13.77 seconds. This is a small production cohort, not an established p95
improvement or proof that storage stalls are fixed.

## Change

The API still runs one in-process Go worker with a PostgreSQL queue, behind ports.
After publishing a thumbnail size, a background job checks for foreground demand
and defers remaining sizes when needed. Its persisted outputs survive the yield;
a later attempt generates only missing sizes. The attempt increment is refunded
under the existing lease fence, without refunding failures or expired work.

Requests and workers also share ownership of each original photo. Readers can
join publication already in progress; workers never wait for that ownership while
holding processing capacity. Cancellation releases ownership and wakes waiting
readers. A yield releases decoded memory; resuming can require another decode.
No additional worker service, queue system, schema migration, CPU, or memory was
introduced for this optimization.

## Reproduction and results

Before: API `f66f2d965`, recorded in the preceding
[diagnosis](thumbnail-regression-diagnosis-2026-09-07.md).
After: API `1123fbad6`, image `sha256:aa3cd385fb506995d44428a5ff007a74a7da85e8c8082cbcfd1cb9c93fd3811d`,
GitOps `a5f8aae`. Both used one worker, 500m CPU, 512Mi memory and the same Garage
storage path. Both followed paced/back-to-back/back-to-back/paced blocks using
the same three originals, immediate small-image opens, queue drains between blocks,
and concurrent warm-image and ordinary asset-list reads at up to two pairs/second.

| Block | Before, photos 1 / 2 / 3 | After, photos 1 / 2 / 3 |
| --- | --- | --- |
| Paced A1 | 4.487 / 3.888 / 3.704 s | 4.120 / 3.606 / 4.034 s |
| Back-to-back B1 | 7.285 / 5.269 / 12.504 s | 3.924 / 4.934 / 4.704 s |
| Back-to-back B2 | 4.588 / 9.506 / 8.792 s | 4.115 / 13.767 / 5.155 s |
| Paced A2 | 3.907 / 4.443 / HTTP 500 at 19.304 s | 3.928 / 3.876 / 4.061 s |

All 12 after-run immediate opens and 494 warm-image reads returned HTTP 200.
Application events matched all 506 image requests: 12 generated and 494 cache hits.
All 494 ordinary reads succeeded. Peak sampled API memory was 341 MiB with no
container restart; increasing memory was unnecessary for this configuration.
Worker logs recorded three cooperative yields and 12 completed jobs.
The queue drained without failed jobs and all four experiment-owned assets were
deleted after measurement.

During back-to-back blocks, concurrent warm reads had median/p95 174/990 ms before
and 125/901 ms after. Ordinary reads had 110/785 ms before and 109/818 ms after.
Those cohorts contain 108 versus 88 samples and are affected by shared production
activity. They do not establish a broad responsiveness improvement. After-run idle
warm reads returned to a 60 ms median and 177 ms p95 (61 samples).

## Trace evidence and limits

For photos 2 and 3 in the back-to-back blocks, gaps after the initial cache miss
were 1.383 / 8.565 / 5.751 / 4.874 seconds before, versus
1.020 / 0.971 / 9.986 / 0.588 seconds after. Foreground generation still took
roughly 3.4–4.0 seconds. This supports reduced typical scheduling interference,
while retaining the counterexample rather than excluding it.

The 13.767-second open spent 9.986 seconds waiting, then 3.389 seconds generating.
Its trace is `e2dae25b2e3c8398b28685f7d5daa07e`. Background root traces were sampled
at 10%, and the holding operation was not captured. An already-running resize or
storage operation cannot yield until its next successful publication checkpoint;
the exact cause of that remaining wait is unresolved. The worker completion log
aligns with the end of the wait (10.251 seconds after request start); it did not
yield during that final running portion of the job.

Independent warm outliers still spent seconds in blob reads with no resizing,
including 12.764 seconds of a 12.896-second request. The earlier direct-Garage
control established stalls outside the API; this scheduler does not repair that
storage path. No physical cause is assigned to SQLite, NFS, or a server here.
The before run also ended on a storage publication timeout, so whole-run latency
aggregates with different idle/drain proportions are not used to claim improvement.

## Validation and release

Scheduler race tests, PostgreSQL claim/refund fencing, cancellation recovery,
authorized HTTP behavior, structural checks and image publication passed
[validation 34080191156](https://github.com/elsell/stuffstash/actions/runs/34080191156).
The code critic found no blocking issue after the cancellation coverage was added.
Full PR checks passed in
[CI 34080576863](https://github.com/elsell/stuffstash/actions/runs/34080576863), including
repeated PostgreSQL search tests, web image, iOS lock, self-host runtime and browser
journey. Builds and tests ran in CI because the local host is disk constrained.
Final PR checks passed in
[CI 34080988756](https://github.com/elsell/stuffstash/actions/runs/34080988756).
[PR 66](https://github.com/elsell/stuffstash/pull/66) merged as `095e5ede9`.
[v0.20.0](https://github.com/elsell/stuffstash/releases/tag/v0.20.0) published signed
API/web images and the self-host bundle. GitOps `8b7b185` deployed API digest
`sha256:63824716a0a0aa2fd4b4e568d93cbd480649d396613710bfbba4763d3fd381c1`
and web digest `sha256:afce9606fc7db23d3fe75c17af547fc34e7258f7c1b8b89e98394977497be66d`.
Both deployments became healthy. The released API runtime source matches the
measured candidate; only tests and documentation changed after measurement.
All 18 released-image thumbnail variants returned 200 and matched reference hashes;
an unauthenticated thumbnail read returned 401. The queue had zero pending,
leased and failed jobs. Private verification is `scheduling-release-smoke.jsonl`.

## Deferred work

- Storage-server/NFS timing and an isolated metadata-storage experiment need the
  previously requested read-only access; no production storage migration was made.
- Preemption inside a resize/storage operation and strict foreground latency bounds
  require further design; cooperative checkpoints do not provide that guarantee.
- Web/mobile telemetry activation, physical-device rendering measurements, broader
  dependency instrumentation, alerts/SLO dashboards and extended profiling remain
  deferred. Existing API telemetry was sufficient for this scheduling comparison.

Private evidence: `scheduling-pacing-abba-1123fbad6.jsonl`, its resource and cache-event
files, and `scheduling-pacing-traces-1123fbad6.json`. The unchanged protocol and
baseline files are retained alongside the after run. No credentials or image bytes
are included in this report.
