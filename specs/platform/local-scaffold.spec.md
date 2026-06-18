# Local Scaffold Spec

## Purpose

Stuff Stash needs a minimal local development scaffold so contributors can run the core service, run tests, and build a container image.

## Scope

This spec covers the first runnable scaffold only:

- A minimal Go core service.
- A health endpoint for local verification.
- Docker-compatible container build files.
- A local Compose file.
- Environment-based runtime configuration.

This spec does not introduce persistence, authentication, authorization, tenancy behavior, REST domain APIs, MCP tools, or mobile app behavior.

## Requirements

- The service must be written in Go.
- Runtime configuration must come from environment variables.
- The service must expose `GET /healthz`.
- `GET /healthz` must return HTTP `200` with a JSON body that identifies the service as healthy.
- The service must not require a database until a persistence spec introduces one.
- The container build must use Red Hat Hardened Images as the base images.
- The build stage must use a pinned Red Hat Hardened Images Go builder image.
- The runtime stage must use a pinned Red Hat Hardened Images core runtime image.
- Base images must be pinned by immutable digest.
- Floating image tags such as `latest` must not be used.
- The container image must run as the hardened runtime image's default user.
- Local Compose must run the app and expose the configured HTTP port.
- Tests must verify the real health endpoint behavior.

## Environment

- `STUFF_STASH_HTTP_ADDR`: address the HTTP server listens on. Defaults to `:8080`.

## Verification

- `go test ./...` must pass.
- `docker compose up --build` should start the app locally.
- `curl http://localhost:8080/healthz` should return a healthy response.

## References

- Red Hat Hardened Images docs describe using a separate Go builder image and core runtime image for multi-stage Go builds.
- This scaffold pins the Go builder image to `registry.access.redhat.com/hi/go:1.25.10-builder-1780418048@sha256:76978661900fd99d3a3d033d736cb9894f317240f52997210ecf0a8e93ce3716`.
- This scaffold pins the runtime image to `registry.access.redhat.com/hi/core-runtime:2.42-1781714135@sha256:50714a55cbfb83cbaaed7f94fffdee8e818cf5b11ac2379ce0ff848513636353`.
