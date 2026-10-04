#!/usr/bin/env python3
"""Derive Astro and preview URLs from GitHub's configured Pages destination."""
import argparse
import json
from pathlib import Path
import re
from urllib.parse import urlsplit


def configuration(pages, preview=None):
    raw = pages['html_url']
    if not isinstance(raw, str) or re.search(r'[\s\x00-\x1f]', raw):
        raise ValueError('Invalid public Pages URL')
    url = urlsplit(raw)
    if url.scheme != 'https' or not url.hostname or url.username or url.password or url.query or url.fragment or url.port:
        raise ValueError('Pages URL must be an HTTPS origin and path')
    if not re.fullmatch(r'/[A-Za-z0-9._/-]*', url.path) or '..' in url.path.split('/'):
        raise ValueError('Invalid Pages base path')
    cname = pages.get('cname') or ''
    if cname and (not re.fullmatch(r'[A-Za-z0-9.-]+', cname) or cname.lower() != url.hostname or url.path != '/'):
        raise ValueError('Custom domain must match the root Pages URL')
    base = url.path.rstrip('/') + '/'
    if preview is not None:
        if not isinstance(preview, int) or preview < 1:
            raise ValueError('Invalid preview pull request number')
        base += f'pr-{preview}/'
    site = f'https://{url.netloc}'
    return {'site': site, 'base': base, 'cname': cname, 'url': site + base}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('pages_json', type=Path)
    parser.add_argument('--preview', type=int)
    parser.add_argument('--env-file', type=Path, required=True)
    parser.add_argument('--output-file', type=Path, required=True)
    args = parser.parse_args()
    values = configuration(json.loads(args.pages_json.read_text()), args.preview)
    with args.env_file.open('a') as output:
        output.write(f"STUFF_STASH_DOCS_SITE={values['site']}\nSTUFF_STASH_DOCS_BASE={values['base']}\n")
    with args.output_file.open('a') as output:
        output.write(''.join(f'{key}={value}\n' for key, value in values.items()))


if __name__ == '__main__':
    main()
