"""Read-only, authenticated thumbnail HTTP comparison. Tokens never enter evidence."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import hashlib
import http.client
import json
import math
import os
import re
import secrets
import ssl
import time
from urllib.parse import urlsplit

PATH = re.compile(r"/tenants/[A-Za-z0-9_-]+/inventories/[A-Za-z0-9_-]+/assets/[A-Za-z0-9_-]+/attachments/[A-Za-z0-9_-]+/thumbnail\?variant=(small|medium|large)$")
MAX_BODY = 16 * 1024 * 1024


def validate_manifest(value):
    origin = value.get("base_url", "")
    parsed = urlsplit(origin)
    if parsed.scheme != "https" or not parsed.hostname or parsed.username or parsed.password or parsed.path or parsed.query or parsed.fragment:
        raise ValueError("Expected an HTTPS origin")
    paths = value.get("paths", [])
    if not isinstance(paths, list) or not 1 <= len(paths) <= 100:
        raise ValueError("Expected 1–100 thumbnail paths")
    targets = []
    for path in paths:
        match = PATH.fullmatch(path) if isinstance(path, str) else None
        if not match:
            raise ValueError("Invalid thumbnail path")
        targets.append((path, match.group(1)))
    if len(set(paths)) != len(paths):
        raise ValueError("Duplicate thumbnail path")
    return origin, targets


def measure(origin, path, variant, ordinal, attempt, token, connection_factory=http.client.HTTPSConnection):
    parsed = urlsplit(origin)
    trace_id = secrets.token_hex(16)
    result = {"variant": variant, "ordinal": ordinal, "attempt": attempt, "trace_id": trace_id,
              "status": 0, "bytes": 0, "outcome": "failure"}
    connection = connection_factory(parsed.hostname, parsed.port or 443, timeout=30, context=ssl.create_default_context())
    started = time.perf_counter()
    try:
        connection.request("GET", path, headers={"Authorization": "Bearer " + token,
            "traceparent": "00-" + trace_id + "-" + secrets.token_hex(8) + "-01"})
        response = connection.getresponse()
        result["headers_ms"] = (time.perf_counter() - started) * 1000
        result["status"] = response.status
        body = response.read(MAX_BODY + 1)
        result["bytes"] = len(body)
        if response.status == 200 and len(body) <= MAX_BODY:
            result["outcome"] = "success"
            result["sha256"] = hashlib.sha256(body).hexdigest()
    except (OSError, http.client.HTTPException):
        # Exceptions may contain request URLs or provider responses; retain no text.
        pass
    finally:
        result["total_ms"] = (time.perf_counter() - started) * 1000
        connection.close()
    return result


def summarize(samples):
    groups = {}
    for sample in samples:
        key = sample["variant"] + (":first" if sample["attempt"] == 1 else ":repeat")
        groups.setdefault(key, []).append(sample)
    summary = {}
    for key, values in groups.items():
        durations = sorted(v["total_ms"] for v in values if v["outcome"] == "success")
        summary[key] = {"requests": len(values), "failures": len(values) - len(durations)}
        if durations:
            summary[key].update({"p50_ms": durations[math.ceil(len(durations) * .5) - 1],
                                 "p95_ms": durations[math.ceil(len(durations) * .95) - 1]})
    return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--revision", required=True)
    parser.add_argument("--repetitions", type=int, default=5, choices=range(1, 31))
    parser.add_argument("--concurrency", type=int, default=1, choices=[1, 2])
    args = parser.parse_args()
    token = os.environ.pop("STUFF_STASH_BENCHMARK_ID_TOKEN", "").strip()
    if not token or any(c in token for c in "\r\n"):
        parser.error("A benchmark token is required in the environment")
    try:
        with open(args.manifest, encoding="utf-8") as source:
            origin, targets = validate_manifest(json.load(source))
    except (OSError, ValueError, TypeError, AttributeError):
        parser.error("Invalid measurement manifest")
    # Reserve evidence before traffic; never overwrite an earlier measurement.
    fd = os.open(args.output, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    samples = []
    with os.fdopen(fd, "w", encoding="utf-8") as output:
        with ThreadPoolExecutor(max_workers=args.concurrency) as pool:
            for attempt in range(1, args.repetitions + 1):
                futures = [pool.submit(measure, origin, path, variant, ordinal, attempt, token)
                           for ordinal, (path, variant) in enumerate(targets, 1)]
                samples.extend(future.result() for future in futures)
                if any(s["status"] in (401, 403) for s in samples):
                    break
        json.dump({"revision": args.revision, "concurrency": args.concurrency,
                   "requested_repetitions": args.repetitions, "samples": samples,
                   "summary": summarize(samples)}, output, indent=2)
        output.write("\n")
    return 0 if all(s["outcome"] == "success" for s in samples) else 1


if __name__ == "__main__":
    raise SystemExit(main())
