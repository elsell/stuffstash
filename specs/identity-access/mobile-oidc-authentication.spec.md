# Mobile OIDC Authentication Spec

## Purpose

Stuff Stash mobile must use the same production SSO boundary as the API and web
app while preserving a low-friction native onboarding experience.

## Scope

This spec covers the first production-shaped mobile authentication flow:

- Provider-neutral OIDC discovery for the mobile app.
- Native-app authorization code with PKCE sign-in.
- Refresh-token backed session renewal.
- Secure mobile token storage.
- Mobile sign-out and session reset.
- Local Dex verification for Expo development builds and physical devices.

This spec does not define account linking, user profile editing, remote push
logout, device management, biometric app lock, offline sync credentials, or
production identity-provider provisioning.

## Decisions

- Mobile authentication must use a system browser or native OIDC session helper
  for authorization. The app must not embed credentials in an in-app web view or
  collect provider passwords directly.
- Mobile must use authorization code with PKCE.
- Mobile must request `openid email profile offline_access` unless an issuer
  rejects `offline_access`; any narrower fallback must keep refresh behavior
  explicit in the resulting session state.
- Mobile must use OIDC discovery metadata instead of hard-coding Dex, Google, or
  provider-specific authorization and token endpoint paths.
- The API is the source of truth for the configured SSO issuer and mobile client
  ID. Mobile must discover those settings from the Stuff Stash instance before
  sign-in instead of compiling Dex or Google values into the app.
- The API must expose only public mobile authentication metadata needed before
  sign-in: issuer, mobile client ID, scopes, and any supported redirect URI
  hints. It must not expose secrets.
- The first mobile slice supports only the app-owned redirect URI
  `stuffstash://auth/callback`. The API must not advertise another mobile
  redirect URI, and the mobile app must fail closed if metadata contains a
  different redirect URI.
- Mobile must send API and realtime voice requests with
  `Authorization: Bearer <id-token>`.
- Refresh tokens, ID tokens, access tokens, expiry timestamps, issuer, subject,
  and client ID are mobile authentication session state and must be stored
  through secure native credential storage.
- The durable connection profile may store only non-secret metadata such as the
  API base URL and selected tenant ID. It must not store tokens.
- Mobile must refresh an expired or near-expired ID token before issuing API or
  realtime voice requests.
- Concurrent requests that need refresh must share one in-flight refresh
  operation rather than racing multiple refresh requests.
- Refresh failure must clear the stored authentication session and return the
  user to sign-in without preserving stale bearer tokens.
- If an already-mounted authenticated mobile surface loses authentication
  because token refresh fails, the API returns an authentication-required
  response, or the configured API authentication mode changes, mobile must clear
  stale secure session state, preserve the non-secret connection profile, and
  return to the sign-in step. The app must surface this through a native
  blocking dialog with a clear `Sign in` action rather than leaving the user on a
  generic `Could not load` error.
- Sign-out must clear secure authentication session state, transient PKCE state,
  selected tenant/inventory preferences, and any in-memory authenticated
  services. Provider end-session support may be added later through discovery
  when it is needed.
- The local-development bearer token path may remain as an explicit development
  fallback, but production mobile authentication must not depend on
  `EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN`.
- Mobile auth code must live behind mobile application ports and adapters.
  Screens may start sign-in or sign-out through application commands, but OIDC
  tokens and provider DTOs must not leak into product UI components.

## Native Runtime

The first mobile implementation uses Expo-compatible native modules:

- `expo-auth-session` for OIDC authorization code with PKCE and token exchange.
- `expo-web-browser` for system-browser session completion.
- `expo-secure-store` for secure token/session storage.

These dependencies must be pinned and recorded in
`specs/platform/tooling-versions.spec.md` before use.

## Instance Authentication Metadata

The API must expose an unauthenticated metadata endpoint suitable for native
clients, such as `GET /.well-known/stuff-stash/mobile-auth` or a versioned
equivalent.

The response must include:

- OIDC issuer URL.
- Mobile OIDC client ID.
- Scopes requested by mobile.
- Supported redirect URI scheme or exact redirect URI patterns when useful.

The response must not include client secrets, local fixture passwords, bearer
tokens, refresh tokens, signing keys, internal network-only endpoints, or
provider-specific private configuration.

If mobile authentication is not configured, the endpoint must fail closed with a
safe error. Local-dev auth mode may return a development-only response only when
the API is explicitly in local-dev mode and the mobile app is explicitly using a
development auth path.

## Onboarding UX

Mobile onboarding must be a guided connection journey:

1. Ask for the Stuff Stash instance URL.
2. Validate the URL and load mobile authentication metadata.
3. Show the configured SSO provider as the next step using provider-neutral copy.
4. Start native sign-in.
5. After successful sign-in, discover available tenants and inventories.
6. Guide empty accounts through tenant and inventory creation.
7. Enter the native tab shell only after both authentication and inventory
   context are ready.

The onboarding screen must use calm product language, clear progress, safe error
states, and one obvious primary action per step. It must not expose OAuth terms,
raw token failures, issuer internals, or local fixture details unless developer
diagnostics are explicitly enabled.

## Security Requirements

- The app must never store OIDC tokens in the connection profile file,
  AsyncStorage, logs, diagnostics, error messages, URL query strings after
  callback handling, or non-secure test fixtures.
- PKCE verifier and state values must be random, single-use, and cleared after
  callback completion or cancellation.
- Callback handling must verify state before storing tokens.
- Token refresh must validate that the refreshed token response contains an ID
  token and a usable expiry.
- Token provider failures must not silently omit the `Authorization` header for
  protected requests; callers must receive an authentication-required result so
  the app can return to sign-in.
- Mobile must not infer authorization from token claims. Tenant, inventory, and
  workflow permissions must continue to come from API responses.
- Realtime voice sessions must use the same authenticated token provider as REST
  requests and must fail closed when no valid token is available.
- Local Dex static users, static clients, and passwords remain local-only
  fixtures and must not be described as a production mobile identity model.

## Local Dex Fixture

Local Dex must include a public mobile client for development builds.

- The mobile client ID must be included in `STUFF_STASH_OIDC_CLIENT_IDS` when
  the API is in OIDC mode.
- The mobile redirect URI must use the configured native app scheme, initially
  `stuffstash://auth/callback`, unless Expo development tooling requires an
  additional generated redirect URI for a specific local validation path.
- Local verification must document how to make the issuer, API, and redirect URI
  reachable from a physical iPhone. The repository must provide a named local
  workflow that renders Dex with a LAN-reachable issuer and starts Compose with
  that same issuer configured for API token verification and mobile metadata.
- Verification scripts may continue to use password grant only as a fixture for
  API-boundary tests. User-facing mobile sign-in must use the native OIDC flow.
- A mobile OIDC verification script must exercise provider discovery,
  authorization code with PKCE, the configured native redirect URI, token
  exchange for the public mobile client, and one refresh-token exchange against
  a reachable local OIDC issuer. The script may drive Dex's local login form as
  a test fixture, but it must not use password grant for the mobile client flow.
  When an API base URL is provided, the script must also verify the API's mobile
  metadata and call a protected API endpoint with the refreshed mobile ID token.

## Verification

- Unit tests must cover mobile auth metadata parsing, missing configuration,
  sign-in state transitions, callback state validation, secure session store
  read/write/clear behavior through fakes, refresh success, refresh failure,
  concurrent refresh coalescing, sign-out cleanup, and token-provider failures.
- Mobile API adapter tests must prove authenticated REST requests use the
  refreshed ID token and fail closed when authentication is unavailable.
- Realtime voice transport tests must prove the WebSocket authorization header
  uses the same token provider and does not connect with an empty or stale token.
- Onboarding tests must prove the app gates tenant/inventory onboarding behind a
  valid authenticated session and returns to sign-in after auth loss.
- Local Dex verification must prove the API accepts an ID token with the mobile
  client audience and rejects wrong-audience tokens.
- Local mobile OIDC verification must prove the mobile public client can obtain
  and refresh an ID token through authorization code with PKCE.
- `pnpm --dir apps/mobile test` and `pnpm --dir apps/mobile check` must pass.
