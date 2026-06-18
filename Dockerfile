# syntax=docker/dockerfile:1.7

ARG GO_BUILDER_IMAGE=registry.access.redhat.com/hi/go:1.25.10-builder-1780418048@sha256:1a99d42f555db97455998945faf3c797c1f65ce1b92e4d9952a589446d114d6c
ARG RUNTIME_IMAGE=registry.access.redhat.com/hi/core-runtime:2.42-1781714135@sha256:82ab1238082f405e19e1cc6e4950549371b6742ba6b649ca356c058249162540

FROM ${GO_BUILDER_IMAGE} AS builder

WORKDIR /src
COPY apps/api/go.mod ./
COPY apps/api/go.sum ./
COPY apps/api/cmd ./cmd
COPY apps/api/internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /tmp/stuff-stash ./cmd/stuff-stash

FROM ${RUNTIME_IMAGE}

COPY --from=builder /tmp/stuff-stash /app/stuff-stash
COPY deploy /deploy
EXPOSE 8080
ENTRYPOINT ["/app/stuff-stash"]
