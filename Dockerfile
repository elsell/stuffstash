# syntax=docker/dockerfile:1.7

ARG GO_BUILDER_IMAGE=registry.access.redhat.com/hi/go:1.25.10-builder-1780418048@sha256:76978661900fd99d3a3d033d736cb9894f317240f52997210ecf0a8e93ce3716
ARG RUNTIME_IMAGE=registry.access.redhat.com/hi/core-runtime:2.42-1781714135@sha256:50714a55cbfb83cbaaed7f94fffdee8e818cf5b11ac2379ce0ff848513636353

FROM ${GO_BUILDER_IMAGE} AS builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /tmp/stuff-stash ./cmd/stuff-stash

FROM ${RUNTIME_IMAGE}

COPY --from=builder /tmp/stuff-stash /app/stuff-stash
EXPOSE 8080
ENTRYPOINT ["/app/stuff-stash"]
