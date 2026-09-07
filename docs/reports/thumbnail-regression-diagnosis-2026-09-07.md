# Thumbnail regression diagnosis

Two problems are present. The application holds shared image-processing capacity
for an entire background job, delaying a newly requested photo behind unfinished
work on an earlier photo. Separately, the Garage storage path intermittently stalls,
including for tiny cached metadata objects and when thumbnail jobs are idle.

The application scheduling mechanism is identified and reproduced. The storage
problem is isolated beyond the API process, but its physical cause—Garage internals,
NFS/server latency, or the storage network—is not yet established. No production
image, deployment configuration, CPU limit, database, or storage layout was changed
during this diagnosis.

## Re-examining the original warm-read result

The one-worker warm p95 of 918 ms was the maximum of only 18 observations under the
nearest-rank calculation. The second-slowest request was 126 ms; the median was
60 ms. Describing that as a general ninefold slowdown overstates what was measured.

Trace `c1ae3ab3c48fa41db372e73e9584669d` accounts for the 918 ms outlier:
842 ms in the first cached blob read, 15 ms in the second, and 895 ms total inside
the API. There was no image generation. The first read retrieves the small
content-type metadata object; the second retrieves the cached JPEG.

The original one-worker immediate cohort showed a different pattern. Five requests
had gaps between the initial cache miss and the next lookup, up to 17.27 seconds.
Those gaps correspond to shared admission/publication waiting in the reader. The
remaining request spent 14.57 seconds writing its generated thumbnail.

## Experiment 1: paced versus back-to-back uploads

The same three source photos were uploaded and immediately opened in four blocks:
paced, back-to-back, back-to-back, paced. Each block started with the thumbnail queue
drained. Paced blocks additionally waited for the queue to drain before each upload.
The API image and one-worker 500m CPU/512Mi configuration stayed fixed. Existing,
cached small images and ordinary asset-list requests were sampled alongside the
experiment. Test-asset deletion was deferred until measurement ended.

| Block | Photo 1 | Photo 2 | Photo 3 |
| --- | ---: | ---: | ---: |
| Paced A1 | 4.487 s | 3.888 s | 3.704 s |
| Back-to-back B1 | 7.285 s | 5.269 s | 12.504 s |
| Back-to-back B2 | 4.588 s | 9.506 s | 8.792 s |
| Paced A2 | 3.907 s | 4.443 s | HTTP 500 after 19.304 s |

The last sample failed on a 15-second thumbnail publication write deadline. The
experiment stopped and cleaned up; there is no successful final idle phase. This
failure is retained and not treated as a successful latency observation.

Traces distinguish queue interference from resizing time:

| Request | Gap after initial cache miss | Generation span |
| --- | ---: | ---: |
| Paced A1, photo 2 | approximately 0 | 3.641 s |
| Paced A1, photo 3 | approximately 0 | 3.501 s |
| Back-to-back B1, photo 3 | 8.565 s | 3.612 s |
| Back-to-back B2, photo 2 | 5.751 s | 3.392 s |
| Back-to-back B2, photo 3 | 4.874 s | 3.505 s |
| Paced A2, photo 2 | approximately 0 | 3.580 s |

Generation spans include publication, so they must not be added to their nested
blob-write spans. A request served by its own background job can legitimately wait
for publication rather than generate in the foreground; that is distinct from
waiting behind an earlier photo.

The worker acquires admission before claiming a job and releases it only after all
requested variants and job resolution finish. Foreground priority changes which
waiter acquires the next permit; it cannot interrupt the current worker. Small-first
publication helps readers of the same photo, but does not let a different photo
interrupt the remaining medium/large work. At two workers, foreground and background
can additionally duplicate work for the same attachment because admission is a
capacity limit, not a per-attachment single-flight mechanism. That duplicate-work
risk follows from the code; this experiment did not independently count duplicate
decodes at concurrency two.

CPU contention is real but does not explain every long delay. During the two
back-to-back blocks the API used approximately 384m and 452m CPU, with 84% and 89%
of cgroup periods throttled. Yet an idle-before cached request also took 14.226 s,
with 14.082 s in a blob read, before test uploads began.

## Experiment 2: bypass the API for a paired storage control

For each pair, an authenticated warm API read ran alongside a signed GET for its
existing 10-byte thumbnail metadata object directly from Garage through a Kubernetes
port-forward. Direct timing measures response headers, bypassing the API's admission,
image processing, authentication and CPU quota. It still includes the diagnostic
tunnel and Garage's storage path; it does not isolate disk from network.

Three phases used the same six existing images: 80 idle pairs, 100 pairs while
uploading three photos for background processing, then 100 recovery pairs. Pairs
started no faster than twice a second and did not overlap previous pairs. Both
idle phases began and ended with zero pending or leased thumbnail jobs. The active
phase also includes time after its three uploads, so it is not pure saturation.

| Phase | API median / p95 | Direct metadata median / p95 | API max | Direct max |
| --- | ---: | ---: | ---: | ---: |
| Idle, 80 pairs | 62.9 / 142.4 ms | 12.0 / 17.5 ms | 269 ms | 22 ms |
| Image work, 100 pairs | 112.6 / 1,329.2 ms | 11.9 / 207.5 ms | 8,491 ms | 5,112 ms |
| Recovery, 100 pairs | 60.2 / 97.2 ms | 11.8 / 15.9 ms | 9,330 ms | 4,262 ms |

In active pair 26, the direct metadata read took 5.112 s and the API read took
5.151 s. Pair 45 took 1.257 s directly and 1.329 s through the API. This reproduces
long storage-path delays outside the API process. Recovery p95 returned near the
original baseline, but isolated stalls remained. Thumbnail generation therefore
aggravates contention; it is not required for every stall.

The API averaged about 66m CPU in the initial control and 277m over the active
phase, with approximately 61% of active-phase cgroup periods throttled. API TCP
counters increased by 10 retransmissions during the active phase and zero in each
idle phase. Those counters do not cover Garage's outgoing traffic and cannot rule
out storage-network problems. Garage's sampled CPU was about 64m during image work,
well below its 500m cap, but one sampled value cannot exclude short bursts.

## Remaining storage investigation

Live configuration and Garage's own statistics confirm SQLite metadata storage.
Both metadata and object data live on the same `nfs-csi` volume served by TrueNAS.
Garage reports ample free space and no block resync errors. Its official deployment
guide recommends fast storage for frequently accessed metadata, ideally an SSD;
that recommendation is context, not evidence that a storage migration will fix
these particular stalls. [Garage deployment guidance](https://github.com/deuxfleurs-org/garage/blob/main-v2/doc/book/cookbook/real-world.md)

The next discriminating measurement is NFS RPC latency and server disk/metadata
latency during a slow paired read, or an isolated comparison using local metadata
storage. Read-only node SSH was unavailable, and the existing node-exporter account
cannot read host NFS mount timing counters. A route/account for those diagnostics
has been requested. No privileged diagnostic workload or production storage migration
was introduced to work around that access limitation.

The application follow-up should address foreground interference and per-attachment
coordination while preserving a single in-process worker system. Any scheduler change
needs its own spec and tests, followed by the same paced/back-to-back experiment.
Increasing worker count or HTTP timeouts would not establish or remove the underlying
storage cause.

Private evidence: `diagnosis-existing-traces.json`, `diagnosis-pacing-abba.jsonl`,
`diagnosis-pacing-traces-initial.json`, and `diagnosis-storage-{control,busy,recovery}.jsonl`.
The private scripts record statuses, trace IDs, phase boundaries, queue snapshots
and cgroup counters. Shared production activity and small cohorts limit generalization;
the conclusions above rely on trace mechanisms and paired controls, not just p95 ratios.
