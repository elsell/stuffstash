# Image-loading performance — September 6, 2026

The production API now uses staged image resizing and the cluster-internal Garage
endpoint. Public uploads retain HTTPS. Both changes were deployed through the
infra GitOps repository; CPU and memory limits were unchanged.

In the controlled comparison, cold-thumbnail median latency fell **45.7%**, from
**6.85 to 3.72 seconds**. Cold p95 fell **28.5%**, from **11.15 to 7.96 seconds**.
Cold loading is still slow, especially for large images.

## Production results

| Thumbnail | Cold median before | Cold median after | Reduction |
| --- | ---: | ---: | ---: |
| Small, 256 px | 6.05 s | 3.02 s | 50.0% |
| Medium, 768 px | 6.82 s | 3.72 s | 45.5% |
| Large, 1600 px | 7.95 s | 5.00 s | 37.1% |

Across all sizes, cached-request median was 77 ms before and 69 ms after; cached
p95 was 130 ms before and 118 ms after. Each run completed 90/90 successful
requests with no API container restarts during the workload.

The same six production photos, approximately 2.06–3.76 MB each, were requested
at all three sizes from `paul` against `https://api.stuffstash.jsksell.com`.
Each run used one request at a time and five passes. Trace-correlated logs confirm
18 generations and 72 cache hits per run. Only the 36 measured derivative and
metadata objects were deleted before candidate runs; all six original objects
were verified present before and after. CPU limit stayed 500m, memory limit
512 MiB, and trace sampling 100% throughout the comparison.

These are API HTTP measurements, **not physical-device rendering times**. The
corpus is small and sequential; p95 across 18 cold requests is the maximum
observed request, not a stable production percentile or an SLO claim. Other
cluster activity was not controlled. Cache-hit improvements are small in absolute
terms. No claim is made that concurrent-load memory pressure or OOMs are fixed.

## What was slow, and what changed

On a cache miss, the API reads the original, decodes/resizes/encodes it, writes the
thumbnail and cache metadata, then returns the image. Those operations are on the
request path. Cache hits read the stored derivative. Concurrent requests for the
same variant share generation within one API process.

The resize change reduces the source in stages before the final Catmull–Rom
filter. A paired codec benchmark on the same CI runner measured 28–34% less time
and 14–43% fewer allocated bytes, depending on size. This synthetic benchmark
isolates codec cost; allocated bytes are not peak resident memory. Three real
medium-size before/after photos were inspected without obvious visual degradation
at thumbnail size; this is a limited visual check, not a broad quality study.

The resize-only production run improved cold median to 4.16 seconds but worsened
p95 to 19.99 seconds. Grafana traces exposed a 16.0-second cache write and
6.6–15.3-second blob reads. API storage traffic was using the public media route.
The second change routes API reads/writes directly to Garage inside the cluster
while independently preserving HTTPS for presigned public uploads.

With both changes, cold p95 was 7.96 seconds. In that slowest trace, generation
still took 5.05 seconds and an original read took 2.57 seconds. The smaller sample
supports the measured improvement; it does not establish that storage tail
latency has been eliminated.

## Mobile, web, and background processing

Mobile passes authenticated thumbnail URLs to native image views: small for
list/map images, medium for detail images, and large for the viewer. Constructing
those references does not itself generate a thumbnail. Web fetches authenticated
image bytes into object URLs, shares in-flight work, and limits thumbnail work to
six concurrent tasks. Both clients benefit from the deployed API improvements;
neither frontend required a release for these fixes.

The API already warms some primary small thumbnails asynchronously after asset
reads. It does not provide a durable upload-time derivative queue, and an image
request can still arrive before warming finishes.

Bounded background generation after upload is a sensible next step: prepare the
likely-to-be-viewed small/medium images before the user opens them, retaining an
authorized on-demand fallback. It needs retry handling, deduplication, deletion
coordination, resource limits, and backfill for existing photos. Moving the same
expensive work into a goroutine alone will not remove its CPU/storage cost. That
worker was deferred to finish and report this measured improvement first.

## Validation and deployment evidence

- Baseline API source: `319dd138a`; resize-only source: `f9a50585d`; final API
  source: `0d84c2da9e9e6dcde66988f749dbc8f4ba8c429e`.
- Final image: `ghcr.io/elsell/stuffstash@sha256:84b5f0e00d1af3c5151ed6ebdac1fc547b569757a261a1709fd90fee93a4ffea`.
- GitOps comparison deployment: `0f374fb`; post-comparison 10% trace sampling:
  `ddf8c4d`. No manual Deployment patches were used.
- [API validation and image build](https://github.com/elsell/stuffstash/actions/runs/34061597496):
  API/blobstore/config/HTTP/bootstrap/observability race tests and API publication
  passed. The overall workflow is red because deferred mobile telemetry type
  checking cannot resolve `node:module` in a test; no frontend was deployed.
- [Paired codec benchmark](https://github.com/elsell/stuffstash/actions/runs/34059380384).
- Code critic reviewed the implementation; confirmed findings were fixed.
  Go formatting and relevant structural checks passed. Builds/tests ran in CI
  because of local disk constraints; the HTTP runner's three Python tests passed.
- [Grafana image-performance dashboard](https://nobleseal2240.grafana.net/d/stuffstash-image-performance).

Raw request samples, safe trace summaries, cache-source logs, codec artifacts and
private visual comparisons are retained in the local sibling directory
`stuffstash-media-measurements`. The reproducible HTTP runner is
`scripts/media/measure_http.py`; its protocol is specified in
`specs/media/image-performance-evidence.spec.md`. Tokens and private request paths
are excluded from committed evidence. Production Dex authentication was obtained
directly; no Infisical benchmark-token setup is needed.

## Explicitly deferred observability work

- Deploying web/mobile request and mounted-image telemetry; physical iOS/Android
  rendering and navigation measurements; the outstanding mobile test type fix.
- Broader database/SpiceDB/provider instrumentation beyond the existing spans
  needed for this investigation.
- Additional dashboards, alerting/SLOs, long-term load/overhead studies and broader
  profiling analysis. Existing API tracing, metrics, logs and profiling remain.
- Capturing all BuildKit/SBOM generator pins for a fully reproducible measurement
  build environment.
