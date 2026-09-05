# Search latency investigation

## Objective

Reduce time from an applied search to usable results without weakening matching,
pagination, tenant isolation, or authorization. Keep mobile server state owned by
TanStack Query and infrastructure behind ports. This investigation follows
`search.spec.md` and the mobile server-state specification.

## Baseline: September 5, 2026

Source baseline: main commit `ceb1cd9439c3856fbdeb5ae89e8bbb1d2f0800fb`.

The mobile `SearchLatency.bench.ts` exercises `SearchAssetsQuery` through the real
inventory adapter with in-memory search and asset reads delayed by 50 ms each.
Inventory discovery is local and photos are absent. It excludes typing debounce,
rendering, image transfers, and real API execution. Results from this Mac:

| Matching assets | Ancestor depth | Search + ancestor reads | Mean completion |
| --- | --- | --- | --- |
| 1 | 0 | 1 + 0 | 51.54 ms |
| 1 | 1 | 1 + 1 | 102.45 ms |
| 1 | 4 | 1 + 4 | 256.16 ms |

The benchmark confirms a client waterfall proportional to ancestry depth.
It does not attribute the user's reported production latency. Re-run using
`pnpm --dir apps/mobile exec vitest bench --run SearchLatency.bench.ts`.

Source inspection also establishes that the GORM search adapter materializes
all candidate asset rows, inventory attachment metadata, open checkouts, and
tenant custom asset types before domain matching and result pagination. Even
subsequent pages repeat candidate matching. This is a scaling concern; backend
timings and row-volume measurements are still required before choosing the
replacement implementation.

## Required evidence before completion

- Backend benchmarks for selective, broad, empty, and paginated searches at
  representative asset and attachment counts, run in CI rather than local builds.
- Authenticated production search measurements when access is available. Public
  health checks alone cannot establish search performance.
- Before/after mobile request counts and latency, including shared and distinct
  ancestor chains, repeated queries, cancellation, and inventory switches.
- Preserve exact and substring semantics across all fields and verify adversarial
  authorization at the real API boundary for any changed search interaction.
- CI tests and build results for the final implementation, plus code critic review.

No optimization has been implemented or claimed complete by this baseline.

## Backend benchmark contract

A dedicated CI job runs the real GORM search port against an ephemeral PostgreSQL
18.1 service pinned to the existing Compose image digest. Fixtures contain 100
and 10,000 assets with two attachment metadata records each. Report latency,
allocations, GORM query count, and rows materialized per search for broad, selective,
empty, attachment-only, and subsequent-page cases. Seed and perform a correctness preflight outside the
timed region; each measured iteration also checks result count and scope to avoid
rewarding broken behavior. The reported latency includes those small checks. This warm repository benchmark
excludes HTTP, authorization, audit writes, and production network latency. Row and
query counters observe GORM Query callbacks; they do not count SQL work internal
to PostgreSQL or any future Raw/Row execution path.
