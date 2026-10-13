#!/usr/bin/env python3
import json
import os
from pathlib import Path
import subprocess
import tempfile
import textwrap


workflow = (Path(__file__).resolve().parents[1] / '.github/workflows/release.yml').read_text()
# Execute the actual index assembly, without a YAML library dependency.
body = workflow.split("          python3 - <<'PYINDEX'\n", 1)[1].split('          PYINDEX', 1)[0]
script = "python3 - <<'PYINDEX'\n" + textwrap.dedent(body) + 'PYINDEX\n'

# Executable registry fake: immutable digest inputs, combined manifests, and metadata.
fake = '''#!/usr/bin/env python3
import json
import os
import sys
from pathlib import Path

args = sys.argv[1:]
if args[:3] == ['buildx', 'imagetools', 'create']:
    assert len(args) == 9, args
    assert args[-2].endswith('@sha256:' + 'a' * 64), args
    assert args[-1].endswith('@sha256:' + 'b' * 64), args
    if os.environ['CASE'] == 'create-fails':
        sys.exit(1)
    Path(args[6]).write_text(json.dumps({
        'containerimage.descriptor': {'digest': 'sha256:' + 'c' * 64},
    }))
elif args[:3] == ['buildx', 'imagetools', 'inspect']:
    architectures = ['amd64'] if os.environ['CASE'] == 'missing-platform' else ['amd64', 'arm64']
    manifests = [{'platform': {'os': 'linux', 'architecture': arch}} for arch in architectures]
    manifests.append({'platform': {'os': 'unknown', 'architecture': 'unknown'}})
    print(json.dumps({'manifests': manifests}))
else:
    sys.exit(2)
'''

for case in ('success', 'missing-metadata', 'bad-digest', 'missing-platform', 'create-fails'):
    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        (root / 'docker').write_text(fake)
        (root / 'docker').chmod(0o755)
        for arch in ('amd64', 'arm64'):
            directory = root / 'image-metadata' / f'release-image-metadata-{arch}'
            directory.mkdir(parents=True)
            for component in ('api', 'web'):
                if case == 'missing-metadata' and arch == 'arm64' and component == 'api':
                    continue
                digest = 'sha256:' + ('a' if arch == 'amd64' else 'b') * 64
                if case == 'bad-digest':
                    digest = 'broken'
                (directory / f'{component}-image-metadata.json').write_text(
                    json.dumps({'containerimage.digest': digest})
                )
        env = dict(
            os.environ, CASE=case, PATH=tmp + ':' + os.environ['PATH'],
            GITHUB_REPOSITORY_OWNER='Elsell', GITHUB_SHA='testcommit',
            GITHUB_OUTPUT=str(root / 'output'),
        )
        result = subprocess.run(
            ['bash', '-e', '-c', script], cwd=tmp, env=env, text=True, capture_output=True,
        )
        assert (result.returncode == 0) == (case == 'success'), (case, result.stderr)
        if case != 'success':
            assert not (root / 'output').exists(), (case, 'published outputs after failure')
        else:
            output = (root / 'output').read_text()
            index_digest = 'sha256:' + 'c' * 64
            assert f'api_image_ref=ghcr.io/elsell/stuffstash@{index_digest}' in output
            assert f'web_image_ref=ghcr.io/elsell/stuffstash-web@{index_digest}' in output
            assert f'api_digest={index_digest}' in output
            assert f'web_digest={index_digest}' in output
        print(case, 'passed')
