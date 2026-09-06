# Image Performance Evidence

## Measurement status

This is an incomplete investigation. Production/device baselines, full runtime
instrumentation, image pipeline changes, and the final comparison remain pending.
Follow `../platform/observability.spec.md` for the delivery contract.

## Synthetic codec baseline — 2026-09-06

Revision: `dc9df5c90` (image processor unchanged from `cd6de4bbf`).
CI run: https://github.com/elsell/stuffstash/actions/runs/34052727411
Artifact: `media-codec-measurements` (14-day CI retention; downloaded separately).
Go 1.25.8, Linux amd64, runner reports AMD EPYC 9V74, benchmark GOMAXPROCS 4.
Input: deterministic 4032 × 3024 RGB pattern, JPEG quality 92, approximately 10.3 MB.
This high-detail synthetic fixture is not a representative household-photo corpus.
Three repetitions of five operations per variant, real production codec, with CPU
and allocation profiling enabled. Storage, authorization, network, and device
rendering are excluded. Keep identical profiling settings in paired comparisons.

| Variant | Mean per-operation range | Allocated bytes per operation | Derivative bytes |
| --- | ---: | ---: | ---: |
| small | 752.8–757.3 ms | 43.77–43.78 MB | 1,401 |
| medium | 800.0–800.8 ms | 95.45 MB | 151,429 |
| large | 906.3–908.6 ms | 183.51–183.52 MB | 882,305 |

These are benchmark averages, not p95/p99 latency or peak resident memory.
The unusually tiny small output reflects this fixture's high-frequency pattern
averaging away during downsampling. Do not generalize its compression ratio.

## Observability verification

- Red test run `34052590822` failed on missing telemetry implementation.
- First adapter run `34052727411` passed telemetry race tests and codec baseline.
- Runtime/configuration red runs `34052799009` and `34052877124` failed before
  implementation.
- Run `34053013594` passed configuration and real OTLP exporter integration tests.
- Code review identified coarse default histogram buckets and unrestricted event
  names; both were corrected. Exporter-failure privacy handling is now covered by real collector failure tests.

## Remaining acceptance gates

- Correlate deployed API traces, metrics, logs, and profiles in Grafana.
- Measure actual HTTP cold/warm loads, upload completion, mobile gallery presentation,
  and web upload behavior with a controlled authorized inventory.
- Measure observability overhead separately from pipeline changes.
- Implement durable derivative generation, shared decoding, concurrency limits,
  mobile gallery loading improvements, and web upload decoupling as measured.
- Repeat the same corpus/workload and report evidence, regressions, and limitations.

## Runtime instrumentation checkpoint

Revision `5d0d22faa` passed CI run `34054029170`: race tests for observability,
configuration, the HTTP adapter (including adversarial telemetry boundary tests),
and bootstrap. API startup now enables OTLP through environment configuration;
HTTP/media scopes and console trace correlation are wired. Collector failures are
counted without exporting response bodies to SDK diagnostics. Streaming interruption
and real WebSocket 101 regressions passed after their red run `34053649692`.
Code critic review of wiring found no new confirmed defect.

Continuous profiling configuration red tests were committed at `a692225fd`; the
profiling adapter and runtime wiring are not implemented yet. Local `go tool pprof`
also attempted compilation and ran out of disk; profile summarization is now in
the CI benchmark workflow. No thumbnail behavior has changed.

## Cluster and ingestion readiness (2026-09-06)

- Read-only access through `paul` and `~/.kube/configs/local-don` succeeded for
  namespace `stuffstash`; API health returned healthy. Flux reported applied
  revision `30efb8086d39e012ecc67634ddfd73e99bc15f53`.
- API image observed at inspection:
  `ghcr.io/elsell/stuffstash@sha256:ddec2c47dcb084296715847dac0ea4f7c4837b83969c905062734ade6eefcc4d`.
  An idle pod snapshot was 3 millicores and 13 MiB; this is not workload evidence.
- OTLP empty protobuf requests from `paul` with Infisical-synced credentials returned
  traces 200, metrics 200, logs 204. This verifies connectivity/authentication, not
  emitted application data, query visibility, or complete signal delivery.
- Profiling endpoint/username are populated, but the synced password is empty;
  the operator has been asked to supply it. An authenticated test inventory is
  still needed for the deployed image workload.
- Go runtime metrics passed race CI `34055169957`; cache response counters passed
  `34055259781`; identity boundary instrumentation passed `34055413925`.
- The observability work was rebased onto main `31ef6edce` to preserve current
  release fixes. No image scheduling behavior has changed yet. The original
  deterministic codec measurements remain synthetic evidence at their recorded
  revision, not a measured production comparison.
