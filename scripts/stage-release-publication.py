#!/usr/bin/env python3
"""Seal already-built assets and publication facts into the retained CI artifact."""
import argparse
import hashlib
import json
from pathlib import Path
import re

parser = argparse.ArgumentParser()
for name in ('directory', 'tag', 'commit', 'api-image', 'web-image'):
    parser.add_argument('--' + name, required=True)
args = parser.parse_args()
root = Path(args.directory)
if not re.fullmatch(r'v\d+\.\d+\.\d+', args.tag) or not re.fullmatch('[0-9a-f]{40}', args.commit):
    raise SystemExit('Invalid release identity')
for image in (args.api_image, args.web_image):
    if not re.fullmatch(r'ghcr\.io/[a-z0-9_./-]+@sha256:[0-9a-f]{64}', image):
        raise SystemExit('Expected immutable GHCR image digest')
assets = [dict(name=path.name, sha256=hashlib.sha256(path.read_bytes()).hexdigest())
          for path in sorted(root.iterdir()) if path.name != 'release-notes.md']
plan = dict(tag=args.tag, commit=args.commit, apiImage=args.api_image, webImage=args.web_image, assets=assets)
(root / 'publication.json').write_text(json.dumps(plan, indent=2) + '\n')
