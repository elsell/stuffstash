# Browser image-observer overhead

Local Chromium run on October 1, 2026; both desktop and Pixel 7 emulation projects
passed. Mobile emulation is not physical-device evidence. The adjacent JSON files
retain all 40 measurements per project, runtime user agent, counts and summary.

The workload uses the production visible-image action and telemetry reporter with
a predecoded 512px synthetic image. Twenty paired measured batches alternate
telemetry enabled/disabled after two warm-up pairs. Each batch covers 20
setup/completion/cleanup lifecycles. All 440 enabled lifecycles reached the controlled
transport with successful outcomes. Percentiles describe batch-average costs,
not individual latency tails. Some measurements approach clock resolution;
these small samples do not support a precise production overhead claim.

Image decode, network delivery, scheduler timing, screen rendering and physical
device performance are excluded. The test checks completeness and valid samples,
not a noisy timing threshold. CI repeats this measurement and retains its own
artifact; the local samples are not substituted for CI evidence.
