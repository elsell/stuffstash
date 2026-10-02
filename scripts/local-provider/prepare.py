#!/usr/bin/env python3
"""Bounded, synthetic-only model setup for the isolated acceptance container."""
import hashlib
import json
import os
from pathlib import Path
import time
import urllib.request

BASE = 'http://127.0.0.1:11434'
MODEL = 'qwen3:0.6b'
DIGEST = '7df6b6e09427a769808717c0a93cadc4ae99ed4eb8bf5ca557c90846becea435'

def request(path, body=None, timeout=30):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(BASE + path, data=data, headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(req, timeout=timeout) as response:
        return json.load(response)

# Verify the tag before download, then verify what the runtime actually loaded.
with urllib.request.urlopen('https://registry.ollama.ai/v2/library/qwen3/manifests/0.6b', timeout=30) as response:
    manifest = response.read(1024 * 1024)
if hashlib.sha256(manifest).hexdigest() != DIGEST:
    raise RuntimeError('Pinned model manifest changed; review before updating')
for attempt in range(30):
    try:
        version = request('/api/version', timeout=2)['version']
        break
    except OSError:
        if attempt == 29:
            raise
        time.sleep(2)
if version != '0.9.5':
    raise RuntimeError('Unexpected runtime version')
request('/api/pull', {'model': MODEL, 'stream': False}, timeout=300)
models = request('/api/tags')['models']
if not any(m['name'] == MODEL and m['digest'] == DIGEST for m in models):
    raise RuntimeError('Downloaded model identity does not match reviewed manifest')
parameters = {'temperature': 0, 'num_ctx': 2048, 'num_predict': 512}
request('/api/create', {'model': 'stuffstash-acceptance', 'from': MODEL, 'parameters': parameters, 'stream': False}, timeout=60)
evidence = {'runtimeVersion': version, 'runtimeImage': os.environ['LOCAL_MODEL_IMAGE'], 'model': MODEL, 'manifestSHA256': DIGEST, 'parameters': parameters, 'syntheticOnly': True}
Path('local-model-evidence').mkdir(exist_ok=True)
Path('local-model-evidence/runtime.json').write_text(json.dumps(evidence, indent=2) + '\n')
