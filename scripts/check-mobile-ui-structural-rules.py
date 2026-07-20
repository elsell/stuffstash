#!/usr/bin/env python3

from __future__ import annotations

import re
import sys
from pathlib import Path


IMPORT_PATTERN = re.compile(
    r"^[ \t]*import\s+(?P<clause>.*?)\s+from\s+['\"]react-native['\"]\s*;",
    re.DOTALL | re.MULTILINE,
)
REQUIRE_PATTERN = re.compile(
    r"\b(?:const|let|var)\s+\{(?P<clause>[^}]*)\}\s*=\s*require\(['\"]react-native['\"]\)",
    re.DOTALL,
)
NAMESPACE_REQUIRE_PATTERN = re.compile(
    r"\b(?:const|let|var)\s+(?P<name>[A-Za-z_$][\w$]*)\s*=\s*require\(['\"]react-native['\"]\)",
)


def source_files(root: Path):
    for path in sorted(root.rglob("*")):
        if path.suffix not in {".ts", ".tsx"} or not path.is_file():
            continue
        if ".test." in path.name or "test-support" in path.parts:
            continue
        if path.as_posix().endswith("/ui/components/AppTextInput.tsx"):
            continue
        yield path


def line_number(source: str, offset: int) -> int:
    return source.count("\n", 0, offset) + 1


def violations(path: Path) -> list[tuple[int, str]]:
    source = path.read_text(encoding="utf-8")
    findings: list[tuple[int, str]] = []

    for match in IMPORT_PATTERN.finditer(source):
        clause = match.group("clause").strip()
        if clause.startswith("type "):
            continue
        value_clause = re.sub(r"\btype\s+TextInput\b", "", clause)
        if re.search(r"\bTextInput\b", value_clause):
            findings.append((line_number(source, match.start()), "value-imports React Native TextInput"))
        for namespace in re.findall(r"\*\s+as\s+([A-Za-z_$][\w$]*)", clause):
            use = re.search(rf"\b{re.escape(namespace)}\.TextInput\b", source)
            if use:
                findings.append((line_number(source, use.start()), "renders TextInput through a React Native namespace"))

    for match in REQUIRE_PATTERN.finditer(source):
        if re.search(r"\bTextInput\b", match.group("clause")):
            findings.append((line_number(source, match.start()), "requires React Native TextInput as a runtime value"))

    for match in NAMESPACE_REQUIRE_PATTERN.finditer(source):
        namespace = match.group("name")
        use = re.search(rf"\b{re.escape(namespace)}\.TextInput\b", source)
        if use:
            findings.append((line_number(source, use.start()), "renders TextInput through a required React Native namespace"))

    return findings


def main() -> int:
    root = Path(sys.argv[1] if len(sys.argv) > 1 else "apps/mobile/src")
    if not root.is_dir():
        print(f"mobile source root does not exist: {root}", file=sys.stderr)
        return 1

    failed = False
    for path in source_files(root):
        for line, reason in violations(path):
            failed = True
            print(
                f"{path}:{line}: {reason}; use the project-owned AppTextInput adapter",
                file=sys.stderr,
            )
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
