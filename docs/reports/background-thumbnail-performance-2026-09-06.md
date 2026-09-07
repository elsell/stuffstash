# Background thumbnail production comparison

Status: experiment in progress; no improvement conclusion yet.

## Method

The reference is API revision `0d84c2da9`, which already includes staged resizing and
internal Garage traffic. The candidate is `392891d17`, adding the durable queue and
in-process workers. Both use the same single API deployment, 500m CPU/512Mi memory
limits, blob storage, authentication and existing telemetry configuration. Candidate
concurrency one and two are separate GitOps configurations of the same image.

The experiment runs from `paul` against the public authenticated API. It copies the
same six JPEG originals into dedicated experiment assets in the benchmark inventory.
The originals are unchanged. Photo sizes are 3,755,993, 2,859,202, 2,790,703,
2,170,153, 2,086,071 and 2,058,024 bytes; private evidence records SHA-256 hashes.

Each configuration receives three batches of six sequential JSON/base64 uploads.
After the final upload in each batch, the driver waits 30 seconds before requesting
all three thumbnail sizes with HTTP concurrency two. A read-only object metadata
probe observes derivative readiness without triggering foreground image generation.
A separate six-image cohort requests the small thumbnail immediately after each
upload. Warm small-thumbnail reads follow the fixed-delay cohort. An authenticated
asset-list request runs roughly once per second throughout the experiment.

The driver records each request's duration, status and trace ID, upload confirmation,
observed readiness, and cohort boundaries. Kubernetes resource metrics and restart
status are sampled alongside it. Readiness is an upper bound affected by the probe
interval and object HEAD latency. Kubernetes metrics are sampled working-set values,
not proof of the instantaneous peak allocation.

This measures API-visible image loading and background preparation, not mobile UI
render time. JSON/base64 upload timing does not substitute for a direct-upload timing
measurement. Automated boundary tests cover direct-upload completion and imports.
The small, fixed corpus cannot establish performance for every phone image size.

## Results

Pending completion of baseline, both candidate runs, and deployment verification.

## Deferred observability

Web and mobile telemetry rollout, physical-device rendering measurements, broader
dependency instrumentation, alerts/SLO dashboards, and extended profiling remain
deferred. Existing API traces, metrics, logs and profiling support this experiment.
