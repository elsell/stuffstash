# Native voice origin correction

The user reported a connection-interrupted failure. Ingress recorded
GET /v1/realtime/voice at 2026-10-03 15:48:55 UTC with HTTP403. Recent API logs
contained no realtime session events. The deployed allowlist contained only the
web frontend origin, while the API is behind TLS termination.

Boundary probes reproduced rejection of the public HTTPS API origin before
authentication. The same request without Origin or with the permitted frontend
origin reached authentication (401). The native-origin explanation is consistent
with React Native's default WebSocket origin behavior; the ingress log does not
itself retain the request Origin.

Infra commit 07396debd7cc8d23d5ec17b4d1b95df829906821 adds only the exact API
origin to the allowlist and updates the deployment configuration checksum.
API image v0.28.26 is unchanged. Server dry-run and required code review passed.
Flux applied the revision and the replacement API replica became ready.

[Before](before.json) reproduces the two failing native-origin cases.
[After](after.json) records eight passing deployed boundary checks: trusted and
absent origins reach authentication, invalid credentials remain unauthorized,
and unknown, duplicate and null origins remain forbidden. No forwarded-header
trust or wildcard origin was introduced.

This proves the origin rejection was corrected. It does not prove authenticated
microphone, transcription, inference or speech playback. The single device
checklist retains that end-to-end check without blocking delivery.
