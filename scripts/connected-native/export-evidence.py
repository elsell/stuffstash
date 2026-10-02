#!/usr/bin/env python3
"""Publish only explicitly named post-auth screenshots and the coarse test stage."""
import json
import os
from pathlib import Path
import shutil
import subprocess

def retain_evidence(exported, evidence):
    allowed = {'connected-safe-persisted-detail': '.png',
               'connected-safe-other-account-setup': '.png', 'connected-stage': '.txt',
               'connected-failure': '.txt'}
    for test in json.loads((exported / 'manifest.json').read_text()):
        for attachment in test['attachments']:
            name = attachment['suggestedHumanReadableName']
            for prefix, extension in allowed.items():
                if name.startswith(prefix + '_') and name.endswith(extension):
                    source = exported / Path(attachment['exportedFileName']).name
                    shutil.copyfile(source, evidence / (prefix + extension))


def main():
    work = Path(os.environ['RUNNER_TEMP']) / 'connected-native'
    evidence = work / 'evidence'
    evidence.mkdir(parents=True, exist_ok=True)
    result = work / 'results.xcresult'
    if result.is_dir():
        exported = work / 'private-attachments'
        try:
            subprocess.run(['xcrun', 'xcresulttool', 'export', 'attachments', '--path', str(result),
                            '--output-path', str(exported)], check=True)
            retain_evidence(exported, evidence)
        finally:
            if exported.is_dir():
                shutil.rmtree(exported)


if __name__ == '__main__':
    main()
