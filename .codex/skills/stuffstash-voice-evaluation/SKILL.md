---
name: stuffstash-voice-evaluation
description: Evaluate Stuff Stash conversational inventory quality from realistic voice traces. Use when Codex is asked to run, inspect, judge, or improve the mobile realtime voice flow, live Gemini voice corpus, agent loop behavior, tool-call traces, action-plan proposals, speech responses, or product-quality pass/fail decisions for home-inventory utterances.
---

# Stuff Stash Voice Evaluation

## Overview

Use this skill to evaluate whether the Stuff Stash voice loop behaves like a useful product, not merely whether deterministic assertions pass. The core job is to preserve full traces, read the actual model/tool behavior, and make a clear pass/fail/needs-follow-up judgment for realistic home-inventory requests.

## Workflow

1. Confirm the relevant spec is current before changing code: `specs/agent-model/mobile-realtime-voice-query.spec.md`.
2. For live Gemini evaluation, set the same environment used by the live corpus before running:

```bash
export STUFF_STASH_GOOGLE_LIVE_TESTS=1
export STUFF_STASH_GOOGLE_CLOUD_PROJECT=<project>
python3 .codex/skills/stuffstash-voice-evaluation/scripts/evaluate_voice_corpus.py --judge codex
```

3. Read the generated run directory, especially `summary.md`, `summary.json`, each `scenarios/*.md` trace, and any `judges/*` files.
4. If `--judge codex` was used, audit the judge reasoning. Do not accept a green verdict unless the explanation fits the trace and rubric.
5. Report scenario-level verdicts and concrete changes needed in prompts, schemas, loop policy, tool metadata, fixtures, mobile UX, or tests.

The harness refuses to report green when no scenarios are extracted or when scenarios are skipped/unknown. Use `--allow-skips` only when intentionally parsing incomplete historical output, and call that out in the final evaluation.

## Harness

The bundled harness writes durable artifacts under `.stuffstash/voice-evals/<timestamp>/` by default:

- `go-test.jsonl`: raw `go test -json` output.
- `go-test-output.txt`: readable test output.
- `summary.json`: machine-readable scenario outcomes and judge results.
- `summary.md`: human-readable run summary.
- `scenarios/*.md`: extracted per-scenario traces.
- `judges/*`: optional Codex-judge outputs.

Useful options:

```bash
# Parse an existing captured run without calling providers again.
python3 .codex/skills/stuffstash-voice-evaluation/scripts/evaluate_voice_corpus.py \
  --input-jsonl .stuffstash/voice-evals/<run>/go-test.jsonl

# Ask Codex CLI to judge traces and fail nonzero on judge follow-up/failure.
python3 .codex/skills/stuffstash-voice-evaluation/scripts/evaluate_voice_corpus.py \
  --judge codex --fail-on-judge
```

## Evaluation Rules

Read `references/rubric.md` before making final quality judgments or when using `--judge codex`.

Evaluate the actual trace, not only the exit code. A product-good pass means the model found or created the right concepts, used authorized tool results, proposed reviewable action plans for writes, paused for approval when needed, and gave clear spoken/display output.

Treat these as failures or required follow-up even if the Go test passed:

- The final response contradicts earlier tool results.
- The loop continues after an action plan should have paused for approval.
- The model asks the user whether to create a clearly requested missing place/container instead of proposing a plan.
- Diagnostics hide the useful tool-call details needed to debug behavior.
- The judge says pass but cites no concrete trace evidence.

Do not replace realistic corpus review with brittle substring checks. Deterministic tests protect invariants; this skill decides whether a Helm-style home user would reasonably trust the voice flow.
