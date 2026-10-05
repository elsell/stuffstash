# CLI pairing activation deadline

The pairing credential response supplies two distinct times: credential expiry and activation deadline. The CLI retains both in project-owned connector registration state and private credential storage. The generated HTTP client remains behind the pairing port.

The server accepts a pending credential only when its activation deadline is strictly later than the current time. Equality is expired. Once a heartbeat activates the credential, the activation deadline no longer limits use of that active credential; credential expiry still applies. Existing saved credentials without this new field remain readable, and worker startup must not reject an already active credential because its activation deadline passed.

The registrar receives time through its injected clock. An exchange with a missing activation deadline is an invalid protocol response. A deadline at or before the current time is rejected before saving or sending a heartbeat. Rotation preserves the previous local credential in this case. Existing credential expiry and exact registration identity checks remain mandatory.

Persistence must still precede activation. The registrar checks the activation deadline again after saving. If the deadline passed during storage, it sends no activation heartbeat and reports that the new credential was saved but the deadline passed, with the normal connector rotation command as recovery guidance. It retains the saved replacement and does not restore or delete credentials: a concurrent worker might already have activated the replacement. This bounded change does not add staged credential storage, compare-and-swap, or new approval commands.

An activation response can be lost or fail after the request starts. Retain the saved replacement and the existing worker-based recovery guidance; never infer that an attempted activation did not occur. Server-side checks remain authoritative for expiry during network transit and clock differences. No client grace period is added.

Critical tests cover missing, expired, exact-boundary, and future deadlines; old credential preservation before save; time passing during save; persistence before activation; generated response mapping; and absence of secrets from output. Existing adversarial pairing HTTP tests remain required.
