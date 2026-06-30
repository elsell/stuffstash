# Stuff Stash Voice Evaluation Rubric

Use this rubric when reviewing live voice corpus traces or Codex-judge output.

## Verdicts

- `pass`: The scenario is product-good for the stated user intent. Minor wording issues are acceptable only when they do not change meaning, hide a required action, or create user confusion.
- `needs_followup`: The scenario completed without a hard failure, but the trace reveals a product problem worth fixing before relying on the behavior broadly.
- `fail`: The scenario produced a wrong answer, unsafe action plan, missing approval pause, hidden provider/session failure, contradictory final response, or a dead-end response where an action plan or clear clarification was expected.

## What To Inspect

Read the full scenario trace, not just the last line. Look for:

- Transcript fidelity: whether the recognized text plausibly matches the spoken request and whether the loop handles likely speech-to-text ambiguity safely.
- Tool selection: whether lookup/search calls are specific enough, authorized, and followed by grounded reasoning.
- Tool result use: whether final answers and plans use visible returned assets instead of guessed IDs, titles, or hidden assumptions.
- Multi-step planning: whether missing but clear destination paths become dependent create commands before a move/create command.
- Approval behavior: whether writes stop at a reviewable plan and do not continue into extra model turns or final speech.
- User value: whether the spoken/display response would help a real home inventory user without requiring them to understand provider internals.
- Diagnostics: whether a developer can reconstruct prompt, tool calls, tool results, model responses, action plans, and repair turns from the trace.

## Common Failure Patterns

- The model finds the target item early, then later says it cannot find it.
- The model says "Do you want me to create it?" for a clear create/move request instead of proposing an approval-backed action plan.
- The model proposes the same action plan repeatedly because the loop did not pause.
- The model drops a missing nested destination segment and moves the item to a broader parent.
- The model invents asset IDs, uses titles as IDs, or uses provider-only command names without normalization.
- The final response is technically safe but product-dead, such as a generic provider issue when the trace contains actionable repair data.

## Required Output From The Evaluating Agent

For each scenario reviewed, provide:

- Scenario name and transcript.
- Deterministic status from the Go test.
- Product verdict: `pass`, `needs_followup`, or `fail`.
- Trace evidence: quote or paraphrase the specific tool/model events that justify the verdict.
- Required changes, if any, tied to the responsible surface: prompt, schema, agent loop, tool catalog, fixture, mobile UX, or deterministic tests.

If a Codex judge was used, explicitly state whether you agree with it and why.
