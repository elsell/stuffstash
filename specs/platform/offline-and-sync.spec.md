# Offline And Sync Spec

## Purpose

Stuff Stash should be clear about offline behavior before mobile work begins.

Offline support is useful, but it adds conflict resolution and sync complexity. It is not required for the initial product.

## Scope

This spec covers the initial offline and sync stance for web and mobile clients.

This spec does not define future offline storage, sync queues, conflict resolution, or local model behavior.

## Requirements

- Offline support is not required initially.
- Users cannot queue inventory actions while offline initially.
- Conversational mode does not need to work offline initially.
- Clients must fail clearly when a network connection is required.
- The system should not pretend an action succeeded when it has not reached the backend.
- Local network self-hosted deployments may work without internet access when all required services are reachable on the user's network.
- Local model support does not imply offline client operation.

## Future Direction

- Future offline support must be specified before implementation.
- Future offline support must define conflict resolution, queued actions, action plan approval timing, audit behavior, and security boundaries.

## Testing

- Initial client tests must verify clear failure behavior for offline or disconnected states once clients exist.
- Future offline features must have tests for queueing, replay, conflicts, authorization, and audit behavior.
