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
  names; both were corrected. Exporter-failure privacy handling remains underway.

## Remaining acceptance gates

- Correlate deployed API traces, metrics, logs, and profiles in Grafana.
- Measure actual HTTP cold/warm loads, upload completion, mobile gallery presentation,
  and web upload behavior with a controlled authorized inventory.
- Measure observability overhead separately from pipeline changes.
- Implement durable derivative generation, shared decoding, concurrency limits,
  mobile gallery loading improvements, and web upload decoupling as measured.
- Repeat the same corpus/workload and report evidence, regressions, and limitations.
