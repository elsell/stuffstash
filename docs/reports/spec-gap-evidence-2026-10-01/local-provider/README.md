# Initial 0.6B local-model acceptance: not passed

Later scoped 4B acceptance passed; see [the October 4 result](instrumented-4b/README.md).
The failure below remains historical evidence.

Run [37071655138](https://github.com/elsell/stuffstash/actions/runs/37071655138),
job [111052254865](https://github.com/elsell/stuffstash/actions/runs/37071655138/job/111052254865),
executed source `969e2ef85273a55da5141fe465a93fc31d3a1df5` on an isolated hosted Linux
runner. The adjacent runtime record contains the verified Ollama image and Qwen3
0.6B manifest digests, plus the derived model's bounded inference parameters.
No tenant data or production credentials were used. The test uses the production
provider factory, profile diagnostic and conversation adapter; its lookup result
is synthetic and does not exercise authenticated inventory tools.

| Requirement | Observed result |
| --- | --- |
| Profile diagnostic calls `ready` with the required status | Failed after 5,182 ms. The diagnostic did not satisfy its contract. The retained evidence does not distinguish malformed output from an incorrect status argument. |
| Lookup, tool-result replay and final answer | Failed after 2,991 ms. The adapter returned one valid `find` call, but its query did not equal the instructed `tent`. The test stopped before supplying the result; final-answer behavior was not reached. |
| Bounded execution and identity | Runtime/model identity verified; both requests returned within the 90-second limit. Container and model-volume cleanup passed. These are failure latencies, not a throughput or successful-response benchmark. |

The first attempt, run37071206631 at `bf72498c`, stopped before inference because
the new test assigned bytes to the domain's JSON string type. That test-only error
was corrected, and compilation now precedes model setup. The second attempt above
is the only completed inference experiment. No unchanged retry is authorized by
this result; no adapter assertion or model expectation was weakened to pass.

This small reference model has not passed Stuff Stash compatibility acceptance.
Do not generalize its result to all local models, to hosted providers, or to the
existing controlled protocol tests. A future run requires a specific changed
model/configuration hypothesis and its own reviewed identity and resource budget.
The workflow is manual so documentation changes cannot trigger repeated inference.
