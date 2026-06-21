# Time Spec

## Purpose

Stuff Stash uses time for audit records, invitation expiry, outbox leases, undo history, and lifecycle behavior.

Time must be deterministic in tests and injectable in application services.

## Requirements

- Application services must use an injected clock port for current time.
- Domain logic must not call `time.Now()` directly when the value affects behavior or persisted state.
- Infrastructure adapters may use real time for adapter-local mechanics only when the value does not affect domain behavior.
- Tests must use fake clocks for expiration, lease, audit timestamp, and undo/redo timestamp behavior.
- The default application clock must use UTC real time.
- Outbox leases, invitation expiration, attachment timestamps, undo timestamps, and audit timestamps must use the injected clock.
