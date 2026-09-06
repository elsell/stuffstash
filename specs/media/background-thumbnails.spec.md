# Durable background thumbnail generation

Status: approved design, implementation in progress.

## Outcome and deployment

Validate and persist image uploads as today, then prepare small, medium and large
thumbnails without holding upload confirmation open for derivative generation.
Use a durable PostgreSQL queue and an in-process Go worker managed by API startup
and shutdown. No separate deployment or message broker is required. Queue,
processing and admission remain injectable ports so the worker can move later.
Cover ordinary uploads, verified direct uploads, imports and existing images.
Keep original blobs, authenticated thumbnail URLs, dimensions and JPEG quality.

## Durable work and lifecycle

Attachment creation and thumbnail scheduling commit in one database transaction,
alongside the existing audit record. Import transactions include the same work.
Only image attachments schedule processing. Queue identities include tenant,
inventory, asset, attachment and derivative revision. Repeated scheduling is
idempotent. Store no user token or signed URL in a job.

Claim work using atomic expiring leases; only the current claim may acknowledge
or retry it. Persist retry eligibility with bounded exponential backoff and retain
exhausted jobs for an explicit operator retry. Use injected time for all deadlines.
Recheck authoritative scoped attachment state before processing and publication.
Coordinate every derivative publisher, including foreground fallback, with
attachment/parent deletion so deletion cleanup cannot finish ahead of a late
publisher that recreates derivatives. Cancel work for deleted attachments; never
reconstruct missing attachment metadata from queued snapshots. Publication and
cleanup coordination must hold across processes and expired/replaced claims.

Generate missing variants small, medium, then large. Reuse source retrieval and
decoding where practical without retaining multiple decoded originals while
waiting for admission. Existing valid derivatives may be reused. Record completion
only after all requested derivatives have been persisted; partial successes may
be reused on retry. Publication failures remain retryable, not successful work.

## Shared resource admission

All cold foreground and background thumbnail generation share one injected,
process-wide admission controller. Acquire capacity before fetching/decoding the
original and release it on every exit. Cache hits do not take capacity.
Concurrency is a positive environment-backed setting; initially compare one and
two, then select the measured default. The capacity counts active images, not
variants or goroutines. Do not spawn unbounded claim batches or decoded buffers.

Waiting foreground work has priority over waiting background work; each priority
is FIFO. Work already running is not preempted. Context cancellation removes a
waiter or releases a concurrently granted permit without leaking capacity.
Releasing a permit twice must be harmless. A waiting worker must stop at shutdown.
Foreground requests retain the existing authenticated on-demand fallback. Share
in-flight work where possible; any residual duplicate processing must still use
the same admission limit. Existing read-triggered warming must use durable work
and cannot bypass shared admission.

## Backfill and operation

Provide resumable, bounded backfill of existing image attachments, maintaining a
durable cursor and idempotent scheduling. Prioritize newly created images over
backfill. Avoid generating already-cached variants. Expose bounded environment
settings for enablement, concurrency, polling, leases, retries and backfill rate.
Disabling workers leaves pending jobs durable and on-demand reads usable.

Use existing injectable observation ports for queue age/depth, processing duration,
variant outcomes and retries/failures. Do not expand unrelated observability.
Operators must be able to inspect and retry exhausted work without raw database
writes or exposing user data in telemetry.

## Validation and rollout

Write failing behavioral tests before implementation. Verify atomic enqueue and
rollback for each creation path, duplicate scheduling, lease recovery and stale
claim rejection, retry/exhaustion, cancellation/shutdown, foreground priority,
shared limits, partial derivatives, backfill resumption and deletion/publication
races. Exercise real HTTP authorization and tenant isolation at affected boundaries,
including legitimate access and unauthenticated, wrong-role and cross-tenant denial.
Use real PostgreSQL for cross-process claim and deletion coordination evidence.

Benchmark one versus two active images on the same photo corpus and resource
limits. Record throughput, memory and ordinary API request responsiveness while
background work is active. Measure upload confirmation, variant readiness,
immediate opening and first opening after readiness separately. Choose two only
if the measurements justify the tradeoff. Deploy exclusively through infra GitOps,
verify health, run the production comparison and report limitations. Code critic
review and relevant CI/structural checks are required before final deployment.


## Claim fencing and attempt accounting

Each acquisition uses a fresh claim token. Resolutions also match the returned
lease deadline, so even accidental token reuse cannot let an expired claimant
resolve replacement work. Canonicalize claim times to UTC microseconds to match
PostgreSQL timestamp precision. Count attempts at acquisition, including work lost
to a crash; explicit failure resolution must not increment that count again.
The worker must not start processing once the configured attempt budget is exceeded.
