#!/usr/bin/env python3
"""Track CLI coverage without embedding the OpenAPI document in the binary."""
import argparse
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
METHODS = {'get', 'post', 'put', 'patch', 'delete', 'head', 'options', 'trace'}


def fingerprint(spec, operation):
    references = {}

    def visit(value):
        if isinstance(value, dict):
            reference = value.get('$ref')
            if reference and reference not in references:
                if not reference.startswith('#/'):
                    raise ValueError('External schema references require an explicit review.')
                target = spec
                for part in reference[2:].split('/'):
                    target = target[part.replace('~1', '/').replace('~0', '~')]
                references[reference] = target
                visit(target)
            for child in value.values():
                visit(child)
        elif isinstance(value, list):
            for child in value:
                visit(child)

    visit(operation)
    canonical = json.dumps({'operation': operation, 'schemas': references}, sort_keys=True, separators=(',', ':'))
    return hashlib.sha256(canonical.encode()).hexdigest()


def inventory(spec):
    result = {}
    for path, item in spec['paths'].items():
        for method, operation in item.items():
            if method not in METHODS:
                continue
            identifier = operation['operationId']
            if identifier in result:
                raise ValueError('Duplicate operation ID: ' + identifier)
            contract = dict(operation)
            contract['parameters'] = item.get('parameters', []) + operation.get('parameters', [])
            contract['servers'] = operation.get('servers', item.get('servers', spec.get('servers', [])))
            contract['security'] = operation.get('security', spec.get('security', []))
            schemes = spec.get('components', {}).get('securitySchemes', {})
            contract['securitySchemes'] = {name: schemes[name] for requirement in contract['security'] for name in requirement}
            result[identifier] = {
                'method': method.upper(), 'path': path,
                'domain': (operation.get('tags') or ['platform'])[0],
                'contractSha256': fingerprint(spec, contract),
            }
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--spec', type=Path, default=ROOT / 'packages/api-client/openapi.json')
    parser.add_argument('--coverage', type=Path, default=ROOT / 'apps/cli/api-coverage.json')
    parser.add_argument('--update', action='store_true')
    parser.add_argument('--require-complete', action='store_true')
    args = parser.parse_args()
    current = inventory(json.loads(args.spec.read_text()))
    saved = json.loads(args.coverage.read_text())['operations'] if args.coverage.exists() else {}
    if args.update:
        entries = {}
        for identifier, contract in sorted(current.items()):
            previous = saved.get(identifier, {})
            same = previous.get('contract') == contract
            entries[identifier] = {
                'contract': contract,
                'commands': previous.get('commands', []),
                'status': previous.get('status', 'pending') if same else 'pending',
                'gaps': previous.get('gaps', ['Command and full request/response workflow coverage are not verified.']) if same else ['Review the current contract and implement the complete command workflow.'],
            }
            if same and previous.get('exclusionReason'):
                entries[identifier]['exclusionReason'] = previous['exclusionReason']
        args.coverage.write_text(json.dumps({'version': 1, 'operations': entries}, indent=2) + '\n')
        saved = entries
    drift = sorted(set(current) ^ set(saved))
    drift += [identifier for identifier in current.keys() & saved.keys() if saved[identifier].get('contract') != current[identifier]]
    if drift:
        parser.exit(1, 'CLI contract coverage needs review: ' + ', '.join(drift[:8]) + '\n')
    incomplete = []
    for identifier, entry in saved.items():
        if entry.get('status') not in {'pending', 'partial', 'implemented', 'excluded'}:
            parser.exit(1, 'Invalid coverage status: ' + identifier + '\n')
        if entry['status'] == 'implemented' and (not entry.get('commands') or entry.get('gaps')):
            parser.exit(1, 'Implemented coverage requires named commands and no recorded gaps: ' + identifier + '\n')
        if entry['status'] == 'excluded' and not entry.get('exclusionReason'):
            parser.exit(1, 'Excluded coverage requires a reason: ' + identifier + '\n')
        if entry['status'] in {'pending', 'partial'}:
            incomplete.append(identifier)
    print(f'{len(current)} API operations tracked; {len(incomplete)} still need implementation or verification.')
    if args.require_complete and incomplete:
        parser.exit(1, 'Full CLI parity is not complete.\n')


if __name__ == '__main__':
    main()
