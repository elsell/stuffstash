# Pinned local-provider round trip: passed

[Run 37173848024](https://github.com/elsell/stuffstash/actions/runs/37173848024)
at `301496f432e6ac03d35428de1ad6a56a8e4408d5` passed against the production
provider factory and compatible conversation adapter. The pinned Ollama 0.9.5
and Qwen3 4B identities and inference parameters are in [runtime.json](runtime.json).
The hosted container retained the two-CPU, 6 GiB limits. Cleanup succeeded.

[Stage evidence](stages.json) records a successful profile diagnostic (42,531 ms),
a `find` call with query `tent` (40,677 ms), then tool-result replay and an `answer`
call preserving the supplied `Garage shelf` location (35,797 ms). Both conversation
responses returned HTTP 200, finish reason `tool_calls`, and exactly one tool call.
The successful assertions establish the synthetic query and answer contents;
provider bodies and arguments are intentionally not retained.

The model, prompt, adapter, prediction budget and successful-output assertions
were unchanged from the prior 4B comparison. This instrumented run separated
lookup and replay into independent 90-second contexts and categorized failures.
Its success does not retrospectively explain the earlier ambiguous lookup failure
or establish repeatability. No unchanged retry is needed.

This closes the narrow real-model lookup/result/answer acceptance requirement for
this pinned configuration. It does not prove authenticated inventory tool execution,
production deployment, speech capture/playback, other models, or general model
quality. Roughly 76 seconds for lookup plus answer is not a claim of acceptable
interactive voice latency. Existing controlled authorization/protocol tests remain
separate evidence.
