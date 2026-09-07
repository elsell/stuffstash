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
is FIFO. A resize or publication already running is not interrupted; background
work checkpoints between successfully published variants as specified below. Context cancellation removes a
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

Runtime integration (API-owned worker goroutines, shared reader/admission, validated
configuration and batch telemetry) passed API race and PostgreSQL jobs in CI
`34066855351`. The same run passed the SQLite WAL publication regression. This is
implementation evidence, not production deployment or performance evidence.
Backfill, ambiguous-write cleanup, operational retry and production comparison
remain required. The follow-up code-critic review resumed and accepted the SQLite fix and
transactional backfill design. PostgreSQL backfill concurrency and rollback tests
remain required.

During rollout, finish replacing API instances that predate transactional scheduling
before enabling the initial backfill. Otherwise an old instance could create an
attachment behind the saved cursor without enqueueing it. Enable backfill in a
subsequent GitOps change after verifying all API replicas use the new image.

Backfill runtime settings use the same prefix: `BACKFILL_ENABLED` defaults false
for the staged rollout; `BACKFILL_BATCH_SIZE` defaults 25 (1–1000); and
`BACKFILL_INTERVAL` defaults 5s (100ms–1h). One API-owned goroutine advances a batch
per interval, stops at completion, and cancels/joins at shutdown. Each transaction
has the lease-duration timeout. A failure waits until the next interval and retries
from the durable cursor. Disabling processing also disables automatic backfill.
The memory adapter provides the same batch semantics within its store lifetime.

The feature CI runs the Go structural hook over changes since the deployed
`0d84c2da9` reference in addition to race tests and PostgreSQL integration tests.
Local compilation remains deferred to CI because workstation disk space is limited.

## Deletion cleanup and late remote writes

Asset deletion must enqueue blob cleanup for its attachments in the same transaction
as the asset deletion and audit, before attachment rows cascade away. Lock the parent
and attachment rows in a consistent order so concurrent attachment creation cannot
escape the deletion snapshot. The memory adapter removes the same metadata and jobs.
Inventory deletion continues to reject any inventory containing assets; do not expand
that product behavior to bulk deletion merely for this feature.

Retain processed blob-deletion outbox records as durable deletion tombstones. Add
bounded, fairly scheduled, leased periodic rechecks of all original and derivative
keys, including metadata. Fenced resolution records the last recheck time without
removing the tombstone or resetting the original processed status. Attempt every
key even if one delete fails. Failures must not starve other due tombstones.
This provides eventual cleanup of remotely accepted writes that finish after a
transport error and initial cleanup. It depends on continued rechecks; it does not
claim to prevent late object creation immediately. Blob storage keys must never be
reused for live attachments after deletion; enforce that creation invariant through
persistence, including imports and direct-upload completion, before enabling rechecks.

A permanent storage-key reservation table enforces non-reuse with a unique key.
Reserve the original key in each attachment/import creation transaction; roll the
reservation back if creation fails and retain it after deletion. Seed reservations
from existing attachments and deletion events in the migration. A failed reserve
must fail creation before any metadata or job commits. This avoids a check-then-insert
race against deletion. The in-memory adapter retains equivalent reservations.
Attachment creation and asset deletion both lock the parent asset before checking
or changing its attachments. Asset deletion snapshots and locks those attachments,
enqueues cleanup using each attachment ID as its deterministic event ID, then deletes
the parent. Preserve the existing explicit attachment-deletion audit behavior.

Rechecks claim one processed deletion at a time with a fresh token and expiring
lease. Eligibility requires the processed timestamp and last recheck to be at least
`CLEANUP_RECHECK_INTERVAL` old (default 1h, bounds 1m–30d). Prefer never-rechecked
records, then oldest recheck time and event ID. Resolution matches token, exact
lease deadline and unexpired lease, records completion time plus a safe failure
flag, and clears the lease. Record attempted recheck time even on failure so other
tombstones can progress. Never replace the original deletion status or error.

One API-owned cleanup goroutine runs independently of thumbnail worker enablement,
uses the configured lease and processing timeout, and polls when idle. Cleanup is
bounded to one original plus six derivative/metadata keys per claim. It does not
acquire image admission because it neither downloads nor decodes image contents.
A cancelled process leaves its lease recoverable. Each resolved recheck emits
`blob_deletion.rechecked` with a safe outcome. Original blob deletion remains on
the existing outbox path; this loop supplies ongoing late-write reconciliation.

Within a recheck, allocate each key at most one eighth of the image cleanup budget,
leaving one share for scheduling and resolution overhead. The overall processing
deadline still applies. A stalled original delete must not consume all derivative
deadlines; verify this with storage that waits until its per-key context expires.

## Operator queue commands

Provide `stuff-stash thumbnail-jobs status` and `stuff-stash thumbnail-jobs retry-failed
--limit N` through an operational port. These local commands require the deployment's
database configuration and privileges, not a public HTTP endpoint. Reject malformed
commands and limits before connecting. Status returns JSON counts for pending,
leased, failed and completed jobs, oldest pending age, and backfill completion; it
must not expose attachment IDs, filenames, tokens or storage keys. Read counts from
a consistent database snapshot. Retry accepts 1–1000 jobs (default 100), locks a
bounded set of failed jobs, clears failure/attempt accounting, and makes them pending
at the injected current time. It leaves pending, leased and completed jobs unchanged.
Emit a safe `thumbnail_jobs.retried` operational event with the retried count.

Operator command logs go to stderr; stdout contains exactly one JSON result. Verify
this through the actual command entrypoint, not only an injected test observer.
SQLite status/retry opens an existing file in read-write mode without automatically
creating a database or running migrations. Schema setup remains an explicit operation.

## Release validation

The API image publication gate runs race tests for all application packages, the
command entrypoint, media domain, affected adapters and bootstrap, plus PostgreSQL
coordination tests. This covers attachment creation and asset deletion outside the
media application package. Deploy with backfill disabled until all API replicas use
the queue-aware revision; then enable the resumable scan in a separate GitOps change.
Compare concurrency one and two against the same image corpus and resource limits,
recording upload confirmation, thumbnail readiness, ordinary API latency, memory
and restarts. Record immediate-open and already-ready reads separately.

For the initial storage-key reservation migration, stop all old API writers before
running migrations and starting the queue-aware API. A rolling overlap allows old
writes after reservation seeding; thumbnail backfill alone cannot repair this gap.
The single-deployment production rollout uses `Recreate`, accepting a brief API
interruption. Keep backfill disabled through this migration/startup step.

## Foreground reads during incremental publication

A foreground reader waiting for shared image capacity must be able to serve its
requested derivative once publication succeeds, even if the background job still
holds capacity for remaining variants. Use an injected thumbnail-readiness port
with one-shot subscriptions keyed by derivative storage key and a publication
notification. The in-process adapter retains only active subscriptions and wakes
all readers of the published key; unrelated keys remain independent.

Subscribe before the initial cache read to avoid a missed-publication race. Keep
one cancellable admission attempt queued to preserve foreground FIFO order. On a
notification, read the requested cached derivative once; if it has disappeared,
continue waiting for admission without spinning. Do not poll storage while waiting,
and do not download or decode the original without admission. Publish the readiness
notification only after blob and metadata publication succeeds. Cache readiness
cancels and joins admission and releases any late permit. Cancellation unsubscribes
and joins the admission attempt. Notifications are hints; stored bytes and normal
authorization remain authoritative. An external processor can replace the readiness
adapter along with worker scheduling when that architecture is introduced.

The earlier cache-poll interval setting is retired. Production traces showed repeated
250ms read timeouts while waiting; publication notifications avoid that added storage
traffic. Verify early publication, failed publication, no polling, independent keys,
unsubscription, late-grant cleanup and cancellation with real port fakes.

Operational catch-up may temporarily increase API CPU using verified spare node
capacity, through GitOps. Restore the reference resource limits before measured
comparisons; do not attribute catch-up performance to the application change or
use it to select the normal worker-concurrency default.

For production comparisons, prefer existing API cache-source events over additional
object-store readiness traffic. Correlate successful thumbnail requests with their
cache/generated events and report request latency. A cache event may follow waiting
for a worker, so it does not establish readiness at request start. Preserve aborted
probe experiments as incomplete evidence; do not include them as completed runs.

## Cooperative scheduling and per-photo coordination

Background processing must stop between successfully published variants when a
foreground admission waiter exists. It must unwind decoded data before releasing
capacity, rather than retain a suspended original while another image is admitted.
Uninterrupted processing still downloads/decodes once; after a cooperative yield,
a later attempt reuses persisted variants and may download/decode the original
again for the remaining sizes. Bounded memory and foreground latency take priority
over decode-once across interruptions. The last variant completes normally.

The admission port exposes a synchronized foreground-demand query. A yield is a
typed operational outcome, not a processing failure. The worker releases its claim
with pending status, no failure, and the configured retry-base cooldown to avoid
reclaim spinning. Fenced persistence refunds that claim's attempt increment exactly
once; real failures and abandoned/crashed leases continue consuming retry budget.
No schema change is needed. Publication/deletion guards remain unchanged.

An injected in-process thumbnail-flight port serializes generation by permanently
reserved original storage key, across all variants and both caller classes. It
provides nonblocking ownership acquisition and a completion signal for existing
ownership, with idempotent release and no historical key retention. Only the owner
may download/decode/generate. Authorization still precedes the reader; coordination
provides neither resource discovery nor authorization.

Foreground readers acquire photo ownership before waiting for global admission;
nonowners wait for requested-variant publication or ownership completion without
holding capacity, then recheck the cache and retry ownership. Subscription precedes
the cache read on every retry to avoid lost publication signals. Cancellation
unsubscribes/releases ownership and joins any admission attempt.

Workers still acquire global capacity before claiming jobs, then try photo ownership
without blocking. A busy photo causes a cooperative deferred resolution, releasing
capacity so its foreground owner can run; workers must never wait for ownership
while holding capacity. Ownership is released on success, failure, cancellation and
yield. Foreground requests need not generate variants they did not request.

Tests must prove inter-photo yielding, persisted small reuse on resume, no refunded
failure/crash attempts, fenced/idempotent yield resolution, same-photo suppression
at concurrency two, cancellation/retry wakeups, and unchanged authorized read and
delete boundaries. Repeat the paced/back-to-back production experiment after CI
and code-critic review; record storage failures separately from scheduling latency.

Repeated PostgreSQL search fixtures must allocate a fresh original storage key for
each attachment creation, matching the permanent reservation contract. Removing
fixture attachment rows must not imply that a previously used blob key is reusable.
