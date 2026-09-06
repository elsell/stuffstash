# Observability and Image Performance Spec

## Purpose

Make mobile and web image latency explainable with correlated traces, metrics,
structured logs, and CPU/allocation profiles, then compare an unchanged baseline
with measured improvements. Observability remains replaceable infrastructure.

## Delivery sequence

1. Instrument the existing behavior without changing thumbnail scheduling.
2. Capture reproducible cold-derivative and warm-derivative baselines, including
   upload completion, list cards, detail galleries, and full-screen photos.
3. Implement the media and frontend improvements specified before each pass.
4. Repeat the same workload and report latency, resource use, failures, and limits.

Do not describe synthetic measurements as deployed or physical-device evidence.
Build and test in CI or an authorized remote environment while the local host
lacks build space. Never alter live inventory solely to create benchmark load.

## Signals and boundaries

- Provide OpenTelemetry traces, metrics, and logs through injected infrastructure
  adapters. Domain and application behavior use project-owned ports; SDK types
  must not enter domain code. Preserve observer fan-out and structured console logs.
- Trace incoming HTTP requests and relevant authentication, authorization, audit,
  blob reads/writes, upload verification, thumbnail lookup/generation, and worker
  operations. Propagate context through ports and HTTP clients. Normalize route
  templates rather than recording raw resource paths or query strings.
- Correlate structured logs with trace/span IDs. Export safe typed events using
  bounded asynchronous batches; telemetry export failures must not fail inventory
  requests or cause an unbounded queue. Flush during bounded graceful shutdown.
- Record HTTP and media-operation latency histograms, request/failure counters,
  derivative cache outcomes, bytes transferred, processing queue depth and active
  tasks, and Go runtime/process resource metrics. Metric dimensions must be bounded
  operation/variant/outcome values, never tenant, asset, attachment, request, or
  principal identifiers.
- Support CPU, allocation, heap, goroutine, mutex, and blocking profiles with
  configurable overhead. Profiling must be disabled by default and stay behind a
  dedicated private operational boundary; never expose pprof on the public API.
  Grafana Cloud profiling uses a replaceable profiling adapter or collector.
- Mobile and web record request and visible-image loading duration, success/failure,
  and surface/variant through existing injectable frontend observability boundaries.
  Keep telemetry delivery credentials out of frontend bundles and runtime responses.
- Never export credentials, headers, raw bodies, photo content, file names, asset
  titles, user speech, provider prompts, or arbitrary error strings. Use explicit
  safe-field allowlists for exported events and sanitize errors by category.

## Configuration and integration

- Environment-backed validated configuration controls service name/version,
  environment, exporter endpoints and authentication, sampling, queue/batch limits,
  timeouts, and profiling. Secrets remain outside source control and artifacts.
- Use pinned reviewed OpenTelemetry SDK/exporter and profiling dependencies;
  document the precise versions before adding them. Use standard protocols and
  keep Grafana-specific setup outside application/domain behavior.
- Grafana dashboards must compare cold and warm paths and show p50/p95/p99 latency,
  cache outcomes, upload verification, failures, and CPU/memory. Link trace details
  to logs/profiles where supported. Provision reusable dashboard definitions.
- Grafana service-account credentials manage Grafana resources. Cloud telemetry
  ingestion requires separate Cloud access-policy credentials and stack endpoints.

## Verification and evidence

- Write behavior tests before implementation with controlled fakes or real in-memory
  SDK exporters. Verify trace context, bounded dimensions, secret exclusion,
  disabled behavior, exporter failure isolation, and shutdown flush.
- Add adversarial HTTP tests before any new operational or telemetry endpoint;
  preserve authentication, authorization, and tenant/inventory isolation.
- Keep benchmark input seed, image dimensions/encoding, concurrency, cache state,
  storage topology, revision, repetitions, and machine resources in the report.
- Record instrumentation overhead before applying optimizations. Compare like for
  like workloads and preserve machine-readable sanitized measurements.
- Run the project code critic after each implementation pass and address findings.

## Initial inspection (2026-09-06)

The baseline revision is `cd6de4bbfd8bdb7bf1737f53f377599cd9a9f3ab`.
Thumbnail misses synchronously read/decode/resize/encode/store; small-primary warming
starts on reads, skips occupied slots, and is not durable. Each variant decodes
separately. Mobile mounts all medium gallery photos. Web upload awaits a small
thumbnail. Direct-upload verification reads the original in both adapter and app.
These are hypotheses to measure, not quantified performance results.

## First adapter slice

- Pin OpenTelemetry Go trace/metric SDK and OTLP HTTP exporters at `v1.39.0`,
  and the matching log SDK/OTLP HTTP exporter at `v0.15.0` (the matching release family already required by the dependency graph).
  These components provide standard OTLP transport without a vendor SDK.
- Introduce a project-owned telemetry port for operation scopes, durations and
  counts. Scopes retain context across application and infrastructure calls.
- Export observer event names with a bounded, explicit field allowlist. Arbitrary
  event messages and raw error text are excluded from remote output.
- Runtime construction accepts injected SDK providers for controlled tests; normal
  construction uses asynchronous OTLP HTTP exporters and environment configuration.
- A manually dispatched media observability CI workflow runs focused Go race tests
  without requiring local compilation; broader CI remains required for delivery.

## Runtime configuration contract

- `STUFF_STASH_TELEMETRY_ENABLED` defaults to false. Disabled mode requires no
  endpoint or credential and starts no exporter workers.
- `OTEL_SERVICE_NAME` defaults to `stuffstash-api`; `OTEL_SERVICE_VERSION` and
  `STUFF_STASH_DEPLOYMENT_ENVIRONMENT` provide non-secret resource identity.
- `OTEL_EXPORTER_OTLP_ENDPOINT` is a validated HTTP(S) base endpoint. Reject URL
  userinfo, query strings, and fragments. `OTEL_EXPORTER_OTLP_HEADERS` contains
  comma-separated URL-escaped key/value pairs; reject malformed or duplicate keys
  and CR/LF without echoing the input in errors.
- `STUFF_STASH_TELEMETRY_SAMPLE_RATIO` defaults to 0.1 and must be finite in [0,1].
  `STUFF_STASH_TELEMETRY_EXPORT_TIMEOUT` defaults to 5s, batch interval to 5s,
  metric interval to 30s, queue capacity to 2048 and batch size to 256. Durations
  and sizes must be positive and batch size must not exceed queue capacity.
- Runtime configuration validation returns fixed messages without secrets.

## Export failure privacy

Exporter decorators must consume terminal export failures after the transport's
bounded retries, increment per-signal dropped-batch counters, and prevent raw
collector bodies from reaching the SDK's global error logger. SDK global handlers
must remain untouched. Shutdown/flush errors return fixed categories only. A clean
shutdown proves resources were stopped; delivery additionally requires zero dropped
batches and evidence at the collector. Failure counters remain readable without
using the failed export path and are exported when metrics delivery recovers.

Before constructing an enabled SDK, reject unsupported nonempty `OTEL_*` variables
with fixed errors. This prevents SDK environment parsers logging raw invalid
values. Supported variables are the base endpoint, headers, service name/version;
project variables own sampling, batching, and timeouts. Validate ambient base
endpoint/headers even when runtime configuration is constructed programmatically.

## HTTP and media adapter instrumentation

Runtime wiring must wrap the API handler with W3C trace-context extraction (without
baggage), preserve response streaming and WebSocket upgrades, and record normalized
route template, bounded method, response status and duration. Never record concrete
paths, query strings, header values or request/response bodies. The existing
console observer includes trace and span IDs when present. Blob/image processor
port decorators measure actual read, write, delete and processing work without
changing cache behavior, authorization or media contents. Wire decorators through
bootstrap, including direct-upload adapters, so the baseline includes both original
reads performed during upload verification.
