#!/usr/bin/env python3
"""Capture and evaluate Stuff Stash live voice corpus traces."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


DEFAULT_TEST_REGEX = "TestGoogleGeminiLiveRealisticVoiceCorpus"
VALID_JUDGE_VERDICTS = {"pass", "fail", "needs_followup"}
VALID_JUDGE_CONFIDENCE = {"low", "medium", "high"}


@dataclass
class Scenario:
    name: str
    outputs: list[str] = field(default_factory=list)
    actions: list[str] = field(default_factory=list)
    elapsed: float | None = None
    transcript: str = ""
    spoken: str = ""
    status: str = "unknown"
    trace_path: str = ""
    judge: dict[str, Any] | None = None

    @property
    def text(self) -> str:
        return "".join(self.outputs)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", default=os.getcwd(), help="Stuff Stash repository root.")
    parser.add_argument("--out", help="Output directory. Defaults to .stuffstash/voice-evals/<timestamp>.")
    parser.add_argument("--scenario-regex", default=DEFAULT_TEST_REGEX, help="Go test -run regex.")
    parser.add_argument("--parent-test", default=DEFAULT_TEST_REGEX, help="Parent Go test name used to extract sub-scenarios.")
    parser.add_argument("--input-jsonl", help="Parse an existing go test -json capture instead of running tests.")
    parser.add_argument("--judge", choices=["none", "codex"], default="none", help="Optional trace judge.")
    parser.add_argument("--codex-bin", default=os.environ.get("CODEX_BIN", "codex"), help="Codex CLI binary.")
    parser.add_argument("--max-judge-scenarios", type=int, default=0, help="Limit judged scenarios; 0 means all.")
    parser.add_argument("--fail-on-judge", action="store_true", help="Exit 2 when a judge returns fail/needs_followup.")
    parser.add_argument("--allow-skips", action="store_true", help="Do not fail when captured scenarios are skipped or unknown.")
    parser.add_argument("--self-test", action="store_true", help="Run parser self-test without external services.")
    args = parser.parse_args()

    repo = Path(args.repo).resolve()
    if args.self_test:
        return run_self_test(repo)

    out_dir = Path(args.out).resolve() if args.out else default_out_dir(repo)
    out_dir.mkdir(parents=True, exist_ok=True)

    jsonl_path = out_dir / "go-test.jsonl"
    output_path = out_dir / "go-test-output.txt"

    if args.input_jsonl:
        source = Path(args.input_jsonl).resolve()
        if source != jsonl_path:
            shutil.copyfile(source, jsonl_path)
        return_code = infer_go_test_exit_code(jsonl_path)
    else:
        return_code = run_go_test(repo, args.scenario_regex, jsonl_path)

    scenarios, package_events = parse_go_test_jsonl(jsonl_path, args.parent_test)
    output_path.write_text(render_plain_output(package_events, scenarios), encoding="utf-8")
    write_scenario_traces(out_dir, scenarios)

    rubric_path = Path(__file__).resolve().parents[1] / "references" / "rubric.md"
    rubric = rubric_path.read_text(encoding="utf-8")
    if args.judge == "codex":
        run_codex_judges(repo, out_dir, args.codex_bin, scenarios, rubric, args.max_judge_scenarios)

    summary = build_summary(return_code, scenarios, out_dir)
    (out_dir / "summary.json").write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    (out_dir / "summary.md").write_text(render_summary_markdown(summary), encoding="utf-8")

    print(f"Voice evaluation artifacts: {out_dir}")
    print(f"Summary: {out_dir / 'summary.md'}")

    if not scenarios:
        print("No live corpus scenarios were extracted; refusing to report a green evaluation.", file=sys.stderr)
        return 3
    unevaluated = [scenario.name for scenario in scenarios.values() if scenario.status in {"skip", "unknown"}]
    if unevaluated and not args.allow_skips:
        print(f"Unevaluated scenarios were captured: {', '.join(unevaluated)}", file=sys.stderr)
        return 4
    if return_code != 0:
        return return_code
    if args.fail_on_judge and any(
        (scenario.judge or {}).get("verdict") in {"fail", "needs_followup"} for scenario in scenarios.values()
    ):
        return 2
    return 0


def default_out_dir(repo: Path) -> Path:
    stamp = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    return repo / ".stuffstash" / "voice-evals" / stamp


def run_go_test(repo: Path, scenario_regex: str, jsonl_path: Path) -> int:
    cmd = ["go", "test", "-json", "./apps/api/internal/app", "-run", scenario_regex, "-count=1", "-v"]
    with jsonl_path.open("w", encoding="utf-8") as sink:
        proc = subprocess.run(cmd, cwd=repo, stdout=sink, stderr=subprocess.STDOUT, text=True, check=False)
    return proc.returncode


def infer_go_test_exit_code(jsonl_path: Path) -> int:
    with jsonl_path.open(encoding="utf-8") as source:
        for line in source:
            if not line.strip():
                continue
            try:
                event = json.loads(line)
            except json.JSONDecodeError:
                continue
            if event.get("Action") == "fail":
                return 1
    return 0


def parse_go_test_jsonl(jsonl_path: Path, parent_test: str) -> tuple[dict[str, Scenario], list[dict[str, Any]]]:
    scenarios: dict[str, Scenario] = {}
    package_events: list[dict[str, Any]] = []
    scenario_prefix = parent_test.rstrip("/") + "/"
    with jsonl_path.open(encoding="utf-8") as source:
        for line in source:
            if not line.strip():
                continue
            try:
                event = json.loads(line)
            except json.JSONDecodeError:
                package_events.append({"Action": "output", "Output": line})
                continue
            test_name = event.get("Test") or ""
            if test_name.startswith(scenario_prefix):
                scenario_name = test_name[len(scenario_prefix) :]
                scenario = scenarios.setdefault(scenario_name, Scenario(name=scenario_name))
                action = event.get("Action")
                if action:
                    scenario.actions.append(action)
                    if action in {"pass", "fail", "skip"}:
                        scenario.status = action
                    if "Elapsed" in event:
                        scenario.elapsed = event.get("Elapsed")
                if "Output" in event:
                    scenario.outputs.append(event["Output"])
            else:
                package_events.append(event)

    for scenario in scenarios.values():
        scenario.transcript = extract_first(scenario.text, r'voice corpus trace for "([^"]+)"')
        scenario.spoken = extract_first(scenario.text, r"(?im)^\s*spoken:\s*(.+)$")
    return scenarios, package_events


def extract_first(text: str, pattern: str) -> str:
    match = re.search(pattern, text)
    return match.group(1).strip() if match else ""


def render_plain_output(package_events: list[dict[str, Any]], scenarios: dict[str, Scenario]) -> str:
    chunks: list[str] = []
    for event in package_events:
        output = event.get("Output")
        if output:
            chunks.append(output)
    for scenario in scenarios.values():
        chunks.append(f"\n===== {scenario.name} ({scenario.status}) =====\n")
        chunks.append(scenario.text)
    return "".join(chunks)


def write_scenario_traces(out_dir: Path, scenarios: dict[str, Scenario]) -> None:
    scenario_dir = out_dir / "scenarios"
    scenario_dir.mkdir(parents=True, exist_ok=True)
    for scenario in scenarios.values():
        slug = slugify(scenario.name)
        path = scenario_dir / f"{slug}.md"
        scenario.trace_path = str(path)
        path.write_text(render_scenario_markdown(scenario), encoding="utf-8")


def render_scenario_markdown(scenario: Scenario) -> str:
    return "\n".join(
        [
            f"# {scenario.name}",
            "",
            f"- Status: `{scenario.status}`",
            f"- Transcript: {scenario.transcript or '(not extracted)'}",
            f"- Spoken: {scenario.spoken or '(not extracted)'}",
            f"- Elapsed: {scenario.elapsed if scenario.elapsed is not None else '(unknown)'}",
            "",
            "## Raw Trace",
            "",
            "```text",
            scenario.text.rstrip(),
            "```",
            "",
        ]
    )


def run_codex_judges(
    repo: Path,
    out_dir: Path,
    codex_bin: str,
    scenarios: dict[str, Scenario],
    rubric: str,
    max_scenarios: int,
) -> None:
    codex_path = resolve_codex_binary(codex_bin)
    judge_dir = out_dir / "judges"
    judge_dir.mkdir(parents=True, exist_ok=True)
    selected = list(scenarios.values())
    if max_scenarios > 0:
        selected = selected[:max_scenarios]
    for scenario in selected:
        prompt = build_judge_prompt(rubric, scenario)
        output_file = judge_dir / f"{slugify(scenario.name)}.json"
        cmd = [
            codex_path,
            "exec",
            "--ephemeral",
            "--sandbox",
            "read-only",
            "--ask-for-approval",
            "never",
            "-C",
            str(repo),
            "--output-last-message",
            str(output_file),
            "-",
        ]
        try:
            proc = subprocess.run(cmd, input=prompt, text=True, capture_output=True, check=False)
        except OSError as exc:
            scenario.judge = {
                "verdict": "needs_followup",
                "confidence": "low",
                "product_quality_notes": [f"Codex judge could not be launched: {exc}"],
                "required_changes": ["Run without --judge or pass --codex-bin with a working Codex CLI path."],
                "trace_citations": [],
                "judgeExitCode": None,
            }
            output_file.write_text(json.dumps(scenario.judge, indent=2, sort_keys=True) + "\n", encoding="utf-8")
            continue
        raw = output_file.read_text(encoding="utf-8") if output_file.exists() else proc.stdout + proc.stderr
        parsed = parse_json_object(raw)
        if parsed is None:
            parsed = {
                "verdict": "needs_followup",
                "confidence": "low",
                "product_quality_notes": ["Codex judge did not return parseable JSON."],
                "required_changes": ["Primary agent must inspect the raw judge output and scenario trace."],
                "trace_citations": [],
                "rawOutput": raw[-4000:],
            }
        parsed = validate_judge_result(parsed, proc.returncode)
        scenario.judge = parsed
        output_file.write_text(json.dumps(parsed, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def resolve_codex_binary(codex_bin: str) -> str:
    if shutil.which(codex_bin):
        return codex_bin
    app_bundle = Path("/Applications/Codex.app/Contents/Resources/codex")
    if codex_bin == "codex" and app_bundle.exists():
        return str(app_bundle)
    return codex_bin


def build_judge_prompt(rubric: str, scenario: Scenario) -> str:
    return f"""You are judging one Stuff Stash voice-agent trace.

Return only JSON with this shape:
{{
  "verdict": "pass|fail|needs_followup",
  "confidence": "low|medium|high",
  "product_quality_notes": ["..."],
  "required_changes": ["..."],
  "trace_citations": ["..."]
}}

Rubric:
{rubric}

Scenario:
- Name: {scenario.name}
- Go status: {scenario.status}
- Transcript: {scenario.transcript or "(not extracted)"}
- Spoken: {scenario.spoken or "(not extracted)"}

Trace:
```text
{scenario.text}
```
"""


def validate_judge_result(result: dict[str, Any], exit_code: int | None) -> dict[str, Any]:
    notes = listify(result.get("product_quality_notes"))
    changes = listify(result.get("required_changes"))
    citations = listify(result.get("trace_citations"))
    verdict = result.get("verdict")
    confidence = result.get("confidence")

    invalid_reasons: list[str] = []
    if verdict not in VALID_JUDGE_VERDICTS:
        invalid_reasons.append("missing or invalid verdict")
    if confidence not in VALID_JUDGE_CONFIDENCE:
        invalid_reasons.append("missing or invalid confidence")
    if not citations:
        invalid_reasons.append("missing trace citations")
    if exit_code not in {0, None}:
        invalid_reasons.append(f"judge exited with code {exit_code}")

    if invalid_reasons:
        verdict = "needs_followup"
        confidence = "low"
        notes.append("Codex judge result was treated as needs_followup: " + ", ".join(invalid_reasons) + ".")
        changes.append("Primary agent must inspect the full scenario trace and judge output before accepting the result.")

    return {
        "verdict": verdict,
        "confidence": confidence,
        "product_quality_notes": notes,
        "required_changes": changes,
        "trace_citations": citations,
        "judgeExitCode": exit_code,
    }


def listify(value: Any) -> list[str]:
    if isinstance(value, list):
        return [str(item) for item in value]
    if value is None:
        return []
    return [str(value)]


def parse_json_object(raw: str) -> dict[str, Any] | None:
    stripped = raw.strip()
    if stripped.startswith("```"):
        stripped = re.sub(r"^```(?:json)?\s*", "", stripped)
        stripped = re.sub(r"\s*```$", "", stripped)
    try:
        value = json.loads(stripped)
    except json.JSONDecodeError:
        match = re.search(r"\{.*\}", stripped, re.DOTALL)
        if not match:
            return None
        try:
            value = json.loads(match.group(0))
        except json.JSONDecodeError:
            return None
    return value if isinstance(value, dict) else None


def build_summary(return_code: int, scenarios: dict[str, Scenario], out_dir: Path) -> dict[str, Any]:
    counts: dict[str, int] = {}
    execution_failures = []
    assertion_failures = []
    skipped_or_unknown = []
    product_followups = []
    product_failures = []
    for scenario in scenarios.values():
        counts[scenario.status] = counts.get(scenario.status, 0) + 1
        if scenario.status == "fail":
            assertion_failures.append(scenario.name)
        if scenario.status in {"skip", "unknown"}:
            skipped_or_unknown.append(scenario.name)
        if scenario.judge:
            verdict = scenario.judge.get("verdict")
            if verdict == "fail":
                product_failures.append(scenario.name)
            if verdict == "needs_followup":
                product_followups.append(scenario.name)
    if return_code != 0 and not assertion_failures:
        execution_failures.append("go test exited nonzero without an extracted scenario assertion failure")
    return {
        "runDirectory": str(out_dir),
        "goTestExitCode": return_code,
        "scenarioCounts": counts,
        "executionFailures": execution_failures,
        "assertionFailures": assertion_failures,
        "skippedOrUnknownScenarios": skipped_or_unknown,
        "productFailures": product_failures,
        "productFollowups": product_followups,
        "scenarios": [
            {
                "name": scenario.name,
                "status": scenario.status,
                "transcript": scenario.transcript,
                "spoken": scenario.spoken,
                "elapsed": scenario.elapsed,
                "tracePath": scenario.trace_path,
                "judge": scenario.judge,
            }
            for scenario in scenarios.values()
        ],
    }


def render_summary_markdown(summary: dict[str, Any]) -> str:
    lines = [
        "# Stuff Stash Voice Evaluation",
        "",
        f"- Run directory: `{summary['runDirectory']}`",
        f"- Go test exit code: `{summary['goTestExitCode']}`",
        f"- Scenario counts: `{summary['scenarioCounts']}`",
        f"- Execution failures: `{summary['executionFailures']}`",
        f"- Assertion failures: `{summary['assertionFailures']}`",
        f"- Skipped or unknown scenarios: `{summary['skippedOrUnknownScenarios']}`",
        f"- Product failures: `{summary['productFailures']}`",
        f"- Product follow-ups: `{summary['productFollowups']}`",
        "",
        "## Scenarios",
        "",
    ]
    for scenario in summary["scenarios"]:
        judge = scenario.get("judge") or {}
        judge_text = ""
        if judge:
            judge_text = f" Judge: `{judge.get('verdict', 'unknown')}` ({judge.get('confidence', 'unknown')})."
        lines.extend(
            [
                f"### {scenario['name']}",
                "",
                f"- Status: `{scenario['status']}`.{judge_text}",
                f"- Transcript: {scenario['transcript'] or '(not extracted)'}",
                f"- Spoken: {scenario['spoken'] or '(not extracted)'}",
                f"- Trace: `{scenario['tracePath']}`",
                "",
            ]
        )
        if judge:
            notes = judge.get("product_quality_notes") or []
            changes = judge.get("required_changes") or []
            if notes:
                lines.append("- Judge notes: " + "; ".join(str(note) for note in notes))
            if changes:
                lines.append("- Required changes: " + "; ".join(str(change) for change in changes))
            lines.append("")
    return "\n".join(lines)


def slugify(value: str) -> str:
    value = re.sub(r"[^a-zA-Z0-9._-]+", "-", value.strip()).strip("-")
    return value[:120] or "scenario"


def run_self_test(repo: Path) -> int:
    with tempfile.TemporaryDirectory(prefix="stuffstash-voice-eval-self-test-") as temp_dir:
        out_dir = Path(temp_dir)
        jsonl = out_dir / "go-test.jsonl"
        events = [
            {"Action": "run", "Package": "x", "Test": "TestGoogleGeminiLiveRealisticVoiceCorpus/move_water"},
            {
                "Action": "output",
                "Package": "x",
                "Test": "TestGoogleGeminiLiveRealisticVoiceCorpus/move_water",
                "Output": 'voice corpus trace for "Move my water bottle to the kitchen."\n',
            },
            {
                "Action": "output",
                "Package": "x",
                "Test": "TestGoogleGeminiLiveRealisticVoiceCorpus/move_water",
                "Output": "spoken: Review this change.\n",
            },
            {"Action": "pass", "Package": "x", "Test": "TestGoogleGeminiLiveRealisticVoiceCorpus/move_water", "Elapsed": 1.2},
        ]
        jsonl.write_text("".join(json.dumps(event) + "\n" for event in events), encoding="utf-8")
        scenarios, _ = parse_go_test_jsonl(jsonl, DEFAULT_TEST_REGEX)
        scenario = scenarios.get("move_water")
        if not scenario or scenario.status != "pass":
            print("self-test failed: scenario status not parsed", file=sys.stderr)
            return 1
        if scenario.transcript != "Move my water bottle to the kitchen.":
            print("self-test failed: transcript not parsed", file=sys.stderr)
            return 1
        if scenario.spoken != "Review this change.":
            print("self-test failed: spoken response not parsed", file=sys.stderr)
            return 1
        if infer_go_test_exit_code(jsonl) != 0:
            print("self-test failed: exit code inference returned failure", file=sys.stderr)
            return 1
        invalid_judge = validate_judge_result({"verdict": "pass", "confidence": "high"}, 0)
        if invalid_judge["verdict"] != "needs_followup":
            print("self-test failed: invalid judge result was accepted", file=sys.stderr)
            return 1
        write_scenario_traces(out_dir, scenarios)
        summary = build_summary(0, scenarios, out_dir)
        (out_dir / "summary.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
        (out_dir / "summary.md").write_text(render_summary_markdown(summary), encoding="utf-8")
    print("self-test passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
