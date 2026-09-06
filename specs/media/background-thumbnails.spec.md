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

## Publication coordination

Use an injectable publication guard shared by foreground and background paths.
The PostgreSQL adapter locks the authoritative attachment row during final blob
and metadata publication, after checking scope and source identity. Background
publication also verifies and locks its current job lease, using attachment-then-job
lock order. Do not hold database locks during decoding or resizing. Apply a bounded
publication deadline and keep network I/O on the internal storage route.

A deletion transaction must acquire the same attachment lock (including through
foreign-key cascade) before it can commit attachment removal and its blob cleanup
outbox. Consequently cleanup cannot become visible until any prior publisher has
finished. A publisher arriving after deletion fails the authoritative lookup and
writes nothing. Verify this ordering against real PostgreSQL with independent
connections, including attachment, asset and inventory deletion. Failed partial
publications remain eligible for the same deletion cleanup.

## Implementation evidence so far

Shared admission behavioral/race tests passed API CI `34064006179`. Initial
transactional enqueue and claim/retry tests passed API CI `34064419806`. Regression
CI `34064490160` demonstrated missing crash attempt accounting before its fix;
lease deadline fencing and acquisition accounting are under validation in
`34064626334`. These checks do not yet prove an executing worker, PostgreSQL
cross-process coordination, backfill, production performance or a deployment.
The background feature remains undeployed until those requirements are verified.

Publication guards receive an injected clock and check lease validity after acquiring
locks, rather than accepting a timestamp captured before a potentially long wait.
A guard must reject empty scope, changed source identity, a deleted attachment,
and an expired or replaced background claim before calling the publisher.


Transaction ownership during publication must be independent of caller cancellation:
reserve the database connection and acquire locks with the caller's bounded context,
then keep the transaction alive until the publication callback actually returns.
Propagate cancellation to blob publication, not to transaction ownership. A cancelled
caller must not let automatic transaction rollback release the lock while a blob
write is still unwinding. Transport errors can leave remotely accepted writes
ambiguous even after the callback exits; durable eventual cleanup for those writes
must be resolved and verified before deployment, rather than claiming that a database
lock alone fences object-storage completion.

## Worker execution

One drain acquires background admission, claims at most one image with a fresh ID,
and processes it under a timeout shorter than its lease. It releases admission on
every exit. Successful processing acknowledges completion; failures retry with
capped exponential backoff until exhaustion. A reclaimed job already beyond the
attempt budget is exhausted without processing. Shutdown cancellation leaves the
claim recoverable instead of acknowledging success. The worker emits a safe
`thumbnail_job.resolved` event after a resolution is persisted.

After claiming, recalculate remaining lease time using the injected clock. Never
process an expired claim. Reserve the configured lease-minus-processing interval
for resolution and reduce processing time when acquisition used part of the lease.
The attempt budget accepts 1–100 acquisitions; exponential backoff is capped.

## Batch image processing

The image adapter exposes batch derivative generation through a port. Validate the
whole requested variant set before work, decode the original once, and publish each
requested variant in small/medium/large order. Keep the existing JPEG settings and
filter. Stop on cancellation or publication failure before producing another
variant. Single-thumbnail generation uses the same implementation, so foreground
and background output remain equivalent. Never require all derivative byte buffers
to remain resident together merely to publish a batch.

Single-thumbnail calls retain the domain convention that an empty variant means
small. Unknown variants are invalid; batch requests require explicit unique sizes.

The claimed-image processor loads attachment metadata through a scoped read port,
verifies that it still matches the job, and checks existing derivative content plus
metadata. It downloads the original only when a variant is missing, requests only
missing sizes from the batch adapter, and publishes each via the lifecycle guard.
It propagates cache read/write failures so the worker retries rather than marking
partial storage work complete. Publication uses a configured bounded context.

The in-memory adapter uses a lifecycle mutex and a separate blob mutex to reproduce
publication/deletion ordering without deadlocking blob callbacks. Its lifecycle
mutex wait is not cancellable; it checks cancellation again before publication.
Use PostgreSQL integration tests, not this adapter, to verify bounded lock waiting
and production shutdown behavior.

## Foreground thumbnail reader

A media application reader accepts an already-authorized scoped attachment and
variant. It checks the shared cache before admission, acquires foreground priority
only on a miss, then checks the cache again in case a worker completed while it
waited. Missing output uses the same batch codec, cache conventions and bounded
publication guard as queued processing, with no background claim. Return generated
bytes directly after successful publication without downloading them again. REST
access checks and read audit remain in the existing authorized application facade.

## Runtime configuration and ownership

The API always constructs and injects the shared foreground reader, even when
background execution is disabled. Worker goroutines start and stop with the API;
cancel and join them before closing repositories. Each loop drains one image at a
time, continues while work is available, and polls on idle or failure. A drain has
a lease-duration timeout. Invalid settings fail startup.

Environment settings (prefix `STUFF_STASH_THUMBNAIL_`): `WORKER_ENABLED` defaults
true; `CONCURRENCY` defaults 1 and accepts 1–8; `POLL_INTERVAL` defaults 1s (100ms–1m);
`LEASE_DURATION` defaults 90s (2s–10m); `PROCESSING_TIMEOUT` defaults 60s (1s–5m);
`PUBLICATION_TIMEOUT` defaults 15s (100ms–1m); `MAX_ATTEMPTS` defaults 5 (1–100);
`RETRY_BASE` defaults 5s (100ms–1h); and `RETRY_MAX` defaults 5m (100ms–24h).
Lease duration must exceed processing timeout, publication timeout must be less
than processing timeout, and retry maximum must be at least its base. The final
production concurrency remains subject to the measured comparison.

SQLite publication acquires its database writer lock with a scoped no-op attachment
ID update before reading authoritative state. This avoids relying on unsupported
row-lock clauses and holds deletion behind publication even in WAL mode. The update
only acquires transaction ownership and does not change attachment identity or
business state. A file-backed WAL concurrency test must verify the ordering.

Batch processing retains the existing thumbnail-generation telemetry operation.
That span includes incremental publication; nested blob spans distinguish storage
latency from image work. Publication errors propagate to the operation result.

A queue-drain error emits `thumbnail_worker.failed` without resource IDs or raw
storage errors. It must not emit a job-resolution event unless resolution persisted.

## Backfill persistence

An operational backfill port advances one bounded batch of attachment IDs in lexical
order. Persist one cursor and completion flag per derivative revision. In a single
transaction, lock that cursor, inspect at most the configured batch size, enqueue
image jobs with backfill priority using attachment/revision conflict-ignore, and
advance the cursor. Existing new-upload jobs keep their priority and state. Scan
non-images as cursor entries without enqueueing them. A batch failure rolls back
both jobs and cursor. New uploads behind a finished cursor are already covered by
the attachment transaction. Concurrent backfill callers serialize on cursor state.
No tenant tokens or blob data enter the cursor. This is an operational cross-tenant
scan behind a repository port, never a user-facing resource discovery endpoint.
