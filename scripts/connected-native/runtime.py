#!/usr/bin/env python3
"""Own an isolated real OIDC/API/SpiceDB stack for one native test command."""
import json
import os
from pathlib import Path
import secrets
import signal
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[2]


def wait_http(url, processes, timeout=90):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if any(process.poll() is not None for process in processes):
            exited = ', '.join(Path(process.args[0]).name for process in processes if process.poll() is not None)
            raise RuntimeError('Connected service exited before readiness: ' + exited)
        try:
            with urllib.request.urlopen(url, timeout=2) as response:
                if response.status == 200:
                    return
        except (OSError, urllib.error.URLError):
            pass
        time.sleep(1)
    raise RuntimeError('Connected service readiness deadline exceeded')


def verify_anonymous_discovery(base_url):
    try:
        with urllib.request.urlopen(base_url + '/me/tenants', timeout=5) as response:
            status = response.status
    except urllib.error.HTTPError as error:
        status = error.code
        error.close()
    if status != 401:
        raise RuntimeError(f'Anonymous tenant discovery returned HTTP {status}, expected 401')
    return status


def stop(processes):
    for process in reversed(processes):
        if process.poll() is None:
            try:
                os.killpg(process.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
    deadline = time.monotonic() + 10
    for process in reversed(processes):
        try:
            process.wait(timeout=max(0.1, deadline - time.monotonic()))
        except subprocess.TimeoutExpired:
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            process.wait()


def main():
    if sys.platform != 'darwin' or os.environ.get('GITHUB_ACTIONS') != 'true':
        raise RuntimeError('Connected native runtime requires a hosted macOS runner')
    if not sys.argv[1:]:
        raise RuntimeError('A native verification command is required')
    def interrupted(_signal, _frame):
        raise KeyboardInterrupt('Connected run terminated')
    signal.signal(signal.SIGTERM, interrupted)
    work = Path(os.environ['RUNNER_TEMP']) / 'connected-native'
    data = work / 'data'
    data.mkdir(mode=0o700)  # Refuse reuse: every journey begins with empty data.
    evidence = work / 'evidence'
    evidence.mkdir()
    dex_source = work / 'source/dex-2dce75009acfcd9df77d57f2d144aae999a4c569'
    config = (ROOT / 'deploy/local/dex/config.yaml').read_text()
    config = config.replace('http://dex:5556/dex', 'http://localhost:5556/dex')
    config = config.replace('0.0.0.0:5556', '127.0.0.1:5556')
    config = config.replace('/srv/dex/web', str(dex_source / 'web'))
    (data / 'dex.yaml').write_text(config)
    key = secrets.token_hex(32)
    env = {**os.environ,
        'STUFF_STASH_HTTP_ADDR': '127.0.0.1:8080',
        'STUFF_STASH_AUTH_MODE': 'oidc', 'STUFF_STASH_AUTHZ_MODE': 'spicedb',
        'STUFF_STASH_INVITATION_PUBLIC_BASE_URL': 'http://localhost:8080/invitations/accept',
        'STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP': 'true',
        'STUFF_STASH_REPOSITORY_MODE': 'sqlite',
        'STUFF_STASH_DATABASE_DSN': str(data / 'inventory.sqlite'),
        'STUFF_STASH_BLOB_STORAGE_MODE': 'filesystem',
        'STUFF_STASH_BLOB_STORAGE_PATH': str(data / 'blobs'),
        'STUFF_STASH_SPICEDB_ENDPOINT': '127.0.0.1:50051',
        'STUFF_STASH_SPICEDB_PRESHARED_KEY': key,
        'STUFF_STASH_SPICEDB_TLS_ENABLED': 'false',
        'STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA': 'true',
        'STUFF_STASH_SPICEDB_SCHEMA_PATH': str(ROOT / 'deploy/spicedb/schema.zed'),
        'STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_INTERVAL': '1s',
        'STUFF_STASH_OIDC_ISSUER': 'http://localhost:5556/dex',
        'STUFF_STASH_OIDC_CLIENT_ID': 'stuff-stash-mobile-local',
        'STUFF_STASH_OIDC_CLIENT_IDS': 'stuff-stash-mobile-local,stuff-stash-local',
        'STUFF_STASH_OIDC_MOBILE_CLIENT_ID': 'stuff-stash-mobile-local',
        'STUFF_STASH_OIDC_MOBILE_REDIRECT_URI': 'stuffstash://auth/callback',
        'STUFF_STASH_OIDC_MOBILE_SCOPES': 'openid,email,profile,offline_access',
    }
    processes, logs = [], []
    def launch(name, *args, environment=env):
        log = (data / (name + '.log')).open('w')
        logs.append(log)
        processes.append(subprocess.Popen([str(work / 'bin' / name), *args],
            cwd=ROOT, env=environment, stdout=log, stderr=log, start_new_session=True))
    result = {'revision': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
              'dex': '2dce75009acfcd9df77d57f2d144aae999a4c569', 'spicedb': '1.47.1',
              'storage': 'SQLite/filesystem', 'stage': 'startup', 'outcome': 'failed'}
    try:
        launch('dex', 'serve', str(data / 'dex.yaml'))
        launch('spicedb', 'serve', '--grpc-addr=127.0.0.1:50051', '--datastore-engine=memory',
               '--http-enabled=false', '--metrics-addr=127.0.0.1:9090',
               environment={**env, 'SPICEDB_GRPC_PRESHARED_KEY': key})
        wait_http('http://localhost:5556/dex/.well-known/openid-configuration', processes)
        deadline = time.monotonic() + 30
        while True:
            try:
                with socket.create_connection(('127.0.0.1', 50051), timeout=1):
                    break
            except OSError:
                if time.monotonic() >= deadline or any(p.poll() is not None for p in processes):
                    raise RuntimeError('SpiceDB readiness failed')
                time.sleep(1)
        launch('stuff-stash')
        wait_http('http://localhost:8080/healthz', processes)
        result['stage'] = 'unauthenticated-access'
        result['anonymousStatus'] = verify_anonymous_discovery('http://localhost:8080')
        result['stage'] = 'native-journey'
        # Keep service secrets out of the XCTest subprocess environment.
        command = subprocess.Popen(sys.argv[1:], cwd=ROOT, start_new_session=True)
        processes.append(command)
        if command.wait(timeout=5400):
            raise RuntimeError('Native journey failed; inspect allowlisted test evidence')
        result.update(outcome='passed', stage='completed')
    finally:
        result['exitedServices'] = [Path(p.args[0]).name for p in processes if p.poll() is not None and p.returncode != 0]
        try:
            stop(processes)
        finally:
            for log in logs:
                log.close()
            (evidence / 'runtime.json').write_text(json.dumps(result, indent=2) + '\n')


if __name__ == '__main__':
    main()
