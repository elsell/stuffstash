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

## PostgreSQL baseline evidence

CI run 33943741006 at commit 324e64955 passed the repository benchmark and
PostgreSQL search-audit test. Three repetitions of five iterations each measured:

| Assets / attachments | Search cases | Latency range | Rows loaded | Allocated bytes/search |
| --- | --- | --- | --- | --- |
| 100 / 200 | all five | 3.43–4.16 ms | 301 | 1.54–1.68 MB |
| 10,000 / 20,000 | all five | 189.25–217.49 ms | 30,001 | 197.48–218.29 MB |

Every case made six GORM queries. Empty and selective queries incur nearly the
same hydration cost as broad queries; subsequent pages repeat it. These are warm
repository numbers from CI, not production API latency. The artifact is
`search-benchmark` on that run. CI exports OpenAPI as a small artifact when the
search response evolves so client code can be generated without a local Go build.

Ancestor-path response enrichment is the first in-progress mobile request-count
optimization. Backend candidate hydration remains to be replaced and measured;
the overall optimization is not complete.

## Backend benchmark contract

A dedicated CI job runs the real GORM search port against an ephemeral PostgreSQL
18.1 service pinned to the existing Compose image digest. Fixtures contain 100
and 10,000 assets with two attachment metadata records each. Report latency,
allocations, GORM query count, and rows materialized per search for broad, selective,
empty, attachment-only, and subsequent-page cases. Seed and perform a correctness preflight outside the
timed region; each measured iteration also checks result count and scope to avoid
rewarding broken behavior. The reported latency includes those small checks. This warm repository benchmark
excludes HTTP, authorization, audit writes, and production network latency. Row and
query counters observe executed GORM Query callbacks (excluding dry-run subquery construction); they do not count SQL work internal
to PostgreSQL or any future Raw/Row execution path.

## Mobile path integration evidence

The generated contract came from CI run 33943961458. With the API path consumed
by the existing TanStack-backed search query, the controlled transport benchmark
measured 52.11–52.34 ms at depths 0, 1, and 4 with one search request and zero
ancestor requests. The simultaneously measured legacy fallback took 52.60,
104.33, and 259.76 ms respectively. Full path correctness and request budgets
are asserted. This excludes backend path resolution and remains synthetic.

Mobile source validation passed 161 files / 1,058 tests, plus TypeScript; the API
client suite passed all 34 tests. The preceding API commit passed required CI
checks, including its breadcrumb security tests. Independent-chain application
benchmarking and final backend measurements remain required.

The path benchmark combines the real PostgreSQL search repository with search
application path enrichment for twenty results at depth four, comparing shared
and distinct ancestor chains. It reports request-local database reads and total
completion time, with fixture updates outside timing. It excludes authentication
and network transport and must verify all twenty complete paths.

## First bounded-candidate result

CI run 33944243251 at 3f84fb550 passed candidate-budget/semantic PostgreSQL tests
and measured 10,000-asset searches at 5.02–5.21 ms broad, 109.73–110.85 ms
selective, 108.72–111.42 ms empty, 47.01–47.62 ms attachment-only, and
5.09–5.53 ms subsequent-page. Broad pages loaded 385 rows and allocated 1.86 MB;
selective queries loaded four rows and empty queries loaded one. These results
justify retaining bounded hydration, but selective/empty cost remains high.
The next measurement separates indexed and conservative candidate branches with
UNION ALL. The earlier query-event counter included dry-run subquery compilation;
subsequent measurements exclude it to report executed queries accurately.

## Benchmark execution bounds

Go stops its test alarm before benchmarking, so CI must enforce a separate
diagnostic QUIT signal at eight minutes, forced termination after another 30
seconds, and a 12-minute job limit. Standard error is included in the retained
benchmark output. PostgreSQL benchmark sessions
have a 10-second statement timeout, tests have a 4-minute timeout, and verbose
output identifies the test/benchmark phase. Always retain completed output.
The Go cache uses the actual `apps/api/go.sum` rather than a nonexistent root
manifest. Run 33944443098 was stopped after over ten minutes without measurement
output; its ordinary checks passed, but no performance result is claimed from it.
