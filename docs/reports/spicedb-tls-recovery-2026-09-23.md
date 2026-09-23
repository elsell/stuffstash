# SpiceDB TLS recovery — September 23, 2026

Expiration notification sweeps repeatedly failed with a downstream gRPC TLS
certificate expiry error. The reported expiry was September 18 at 19:12:50 UTC.
API health checks continued succeeding during the outage.

## Evidence and recovery

- The deployed CA and leaf secrets had already renewed on August 19 and expire
  November 17. Both certificate resources reported Ready, revision 2.
- The API pod's mounted CA matched the renewed CA. The live SpiceDB server
  presented the renewed leaf certificate.
- SpiceDB's own CheckPermission server logs returned the same downstream TLS
  error to the API. This locates the failure beyond the API's initial connection.
- The SpiceDB process had been running since before renewal. Stale in-process
  dispatch TLS state is the leading explanation; the exact cached certificate
  or trust object was not inspected.
- A rolling restart of deployment `stuffstash-spicedb-spicedb` in namespace
  `stuffstash` completed successfully. No certificate, credential, data, or TLS
  verification settings were changed.
- From 00:40:45 UTC, the verification sample contained 14 successful
  CheckPermission calls and zero `notification_worker.failed` events. The new
  SpiceDB pod was Ready. Actual notification receipt on a phone was not tested.

## Remaining prevention work

The restart restores service but does not resolve rotation lifecycle handling.
Track renewal propagation through SpiceDB server and dispatch clients, the API's
startup-loaded CA pool, and its init-container trust bundle. Design and test a
rotation procedure with overlapping trust and explicit reload/restart behavior;
do not assume a Ready Certificate means every consumer has reloaded it.
Worker failure alerting must detect this independently of HTTP health checks.

The mobile audit remains open. The most recent verified TestFlight release is
0.24.27 (116.2); this operational recovery did not create a mobile release.
