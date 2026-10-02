# Local-model capacity comparison

The single larger-model comparison did not pass overall. Run
[37078513169](https://github.com/elsell/stuffstash/actions/runs/37078513169)
at `3d0ec4917666f61a0b05447b165eedeaa0049369` compiled the production adapter
test, verified the pinned Qwen3 4B model, performed real inference and removed
the runtime and model volume.

- Profile diagnostic passed in 24,886 ms.
- Lookup acceptance failed in 47,558 ms at the required lookup assertion.
  That assertion combines provider error, call count and tool name; the retained
  evidence cannot distinguish which condition failed.
- Tool-result replay and the final answer were not reached.

The larger configuration improves the observed diagnostic result, but does not
prove an end-to-end local conversation. Do not generalize this failure to every
local model. The unchanged assertions and original 0.6B failure remain retained.
The comparison budget is exhausted: no further unchanged runs or model-tuning
loop. Local round-trip acceptance stays open.

[runtime.json](runtime.json) records the pinned identity and inference settings;
[stages.json](stages.json) records the two measured outcomes. The container was
limited to two CPUs and 6 GiB memory, with 90-second request timeouts. These are
single synthetic measurements, not a production performance benchmark.
