---
title: Compatible Language Providers
description: Use a hosted or local Chat Completions model with Stuff Stash.
---

Stuff Stash can run its inventory conversation through an OpenAI-compatible
Chat Completions endpoint. The model must support function calling. Inventory
permissions and approval of changes remain enforced by Stuff Stash.

## Allow an Endpoint

Set `STUFF_STASH_COMPATIBLE_PROVIDER_ENDPOINTS` on the API to the exact API base
URL, including its path. For example, a local runtime may expose
`http://model-server:8080/v1`. The adapter appends `/chat/completions`.
Multiple base URLs can be separated by commas. Restart the API after changing
the allowlist. Choose operator-controlled hosts; this setting grants network
access to those destinations. Redirects are rejected.

Remote `openai_compatible` profiles require HTTPS. `local_http` profiles may use
HTTP for an explicitly allowed endpoint on your own network. Endpoint and model
names have no implicit defaults.

## Create and Test a Profile

As a household administrator, create a **Language inference** provider profile:

1. Choose **OpenAI compatible** or **Local HTTP**.
2. Enter the same base URL the operator allowed and a model the endpoint serves.
3. Save an API key or OAuth bearer credential through the profile's credential
   action. For a local runtime that ignores authentication, save a non-secret
   marker; the adapter still sends it as a Bearer value.
4. Run the provider test, which sends a short synthetic function-call request.
5. Select the tested profile as the language provider in your voice configuration
   or conversation workflow.

An optional runtime setting such as `{"httpTimeout":"60s"}` controls the request
timeout. It uses Go duration syntax and must be positive. Requests and responses
are each limited to 1 MiB. Provider errors and malformed tool output fail safely.

## Capability Limits

This adapter provides language inference. Keep separate speech recognition and
speech synthesis profiles for voice sessions. A runtime that serves plain chat
but cannot return function calls will fail the diagnostic. Protocol compatibility
does not establish a model's inventory reasoning quality; test your normal
read and approval workflows with the chosen model before relying on it.

Stuff Stash uses the [Chat Completions function-calling contract](https://developers.openai.com/api/docs/guides/function-calling).
It does not require provider-hosted tools or automatically execute model output.
