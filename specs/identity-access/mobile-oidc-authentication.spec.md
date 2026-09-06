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
  return to the combined connection/sign-in screen with the server address
  prefilled. The app must surface this through a native
  blocking dialog with a clear `Sign in` action rather than leaving the user on a
  generic `Could not load` error.
- User-initiated `Sign out` must clear secure authentication session state,
  transient PKCE state, session-only inventory selection, and any in-memory
  authenticated services while preserving the non-secret saved server URL and
  tenant hint. It must return the user to the combined connection/sign-in screen
  with that server prefilled.
- `Change server` is a separate operation. It must sign out, clear the durable
  non-secret connection profile including the server URL and tenant hint, and
  return the user to the combined connection/sign-in screen with no saved server
  address.
- Neither `Sign out` nor `Change server` deletes server-side tenant, inventory,
  asset, provider-profile, audit, or identity data. Mobile confirmation copy
  must name the operation and the local state it clears; a generic `Continue`
  action is not sufficient for changing servers.
- Provider end-session support may be added later through discovery when it is
  needed.
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

The approved mobile flow combines server connection and browser sign-in on one
screen, followed only by the setup the authenticated account actually needs.
This replaces the separate Instance, Sign in, Tenant, and Inventory screen
sequence. Internal application states may remain distinct; they must not force
separate user-facing screens.

### Connection and sign-in

- Heading: `Connect to Stuff Stash`.
- Field label: `Server address`, with an example placeholder such as
  `https://stash.example.com`. Examples are not runtime defaults.
- Primary action: `Connect and sign in`.
- The short note beside the primary action is `Your browser will open for
  sign-in, then bring you back here.`
- Do not add an introductory subtitle or always-visible server-address helper
  sentence. Provide the secondary text action `Need help connecting?`, which
  expands inline help only on request. Help explains that the app needs a running
  Stuff Stash server and its full address, including any required port or path;
  someone joining another person's inventory can request that server address.
- One tap validates and normalizes the address, loads mobile authentication
  metadata, saves non-secret connection metadata, and starts the existing native
  authorization-code-with-PKCE browser flow. There is no intermediate SSO screen,
  provider-explanation card, or second sign-in confirmation button.
- Submitting a different server address must apply existing server-change
  cleanup before authentication or discovery for the new destination. Do not
  reuse credentials, tenant hints, or authenticated services across servers.
- Address and connection errors remain on this screen. Browser cancellation
  returns here with the entered address preserved and a usable retry action;
  cancellation does not establish a session or create tenant/inventory data.
- During an active operation, show loading feedback in the primary button and
  prevent duplicate submissions. Ignore late results from an abandoned flow.
- After successful authentication, discover authorized tenants and inventories.
  A returning user with usable inventory context proceeds directly to the native
  tab shell. Do not create new resources merely because onboarding was opened.

### Household and inventory setup

`Household` is the onboarding presentation of the existing tenant concept. It
introduces no new domain entity or authorization boundary and does not change
support for organization tenants.

| Authenticated account state | Screen | Fields | Primary action |
| --- | --- | --- | --- |
| No usable tenant, eligible for creation under existing policy | `Set up your household` | `Household name`; `First inventory` | `Create household` |
| Existing usable tenant with inventory-creation permission, but no inventory | `Create your first inventory` | `Inventory name` | `Create inventory` |
| Usable inventory already exists | No setup screen | None | Enter the tab shell |

- Household name starts empty, with an example such as `e.g. Maple Street
  household`. The inventory name starts with the editable value `Home Inventory`.
  Both creation screens omit redundant subtitles and inventory helper paragraphs.
- `Create household` collects both names before executing tenant creation and
  first-inventory creation through the existing application services and ports.
  Validate both names before either write; preserve authorization, tenancy, and
  audit history for each command. A combined screen does not imply an atomic
  multi-request transaction or justify bypassing domain services.
- If tenant creation succeeds but inventory creation fails, retain the created
  tenant context and entered inventory name, explain the remaining failure, and
  retry inventory creation without creating another tenant. Reconcile uncertain
  write outcomes before retrying; never blindly repeat creation or silently
  delete already-created resources. If discovery cannot confirm an uncertain
  write, keep retry read-only and ask the user to check its status; do not resend
  creation automatically. Relaunch discovery must reuse that tenant.
- Users without usable context or permission must receive the existing safe
  access/error state; the streamlined flow must not grant creation permission.
- Enter the native tab shell only after authentication and usable inventory
  context are ready. Do not introduce a standalone success/tutorial screen; the
  prototype completion marker is a review aid, not a product screen.

### Start over

- Both setup screens expose `Sign out and start over` immediately below their
  primary action. It is a full-width ghost button with centered text, no border
  or filled resting background, and the same minimum height, corner radius, and
  type size/weight as the primary button. Use subtle hover/pressed feedback and
  visible keyboard focus.
- This explicit action applies the existing change-server cleanup: clear secure
  authentication and transient PKCE state, saved connection/tenant metadata,
  session-only selection, authenticated services, and draft setup fields. Return
  to connection with an empty address and restore the editable default inventory
  name. It must invalidate pending callbacks/results from the abandoned flow.
- It does not delete anything already created on the server. If a previous
  partially completed setup created a tenant, subsequent discovery reuses it.
- Ordinary Settings `Sign out` still preserves the saved server and tenant hint;
  the onboarding start-over action is explicitly broader.

### Presentation and language

- Use the existing brand mark, semantic light/dark colors, and native system
  typography. Keep the brand and heading at consistent top positions rather
  than vertically centering a variable-height block. Use a strong heading,
  regular-weight supporting text, and clear field labels; avoid making all text
  heavy. Do not show a global four-step progress indicator for this branching flow.
- At default text size, reference geometry is a 24-point horizontal inset,
  54-point minimum primary/ghost button height, 12-point button radius, and
  10-point gap between stacked primary and ghost actions. Actions sit near the
  bottom, respect native safe areas, and remain reachable above the keyboard or
  through scrolling. These are minimums, not clipping constraints: Dynamic Type,
  small screens, long translations, and landscape may expand/reflow content.
- Every field uses the shared mobile input primitive. Address entry disables
  autocorrection/capitalization and uses the URL keyboard. Provide accessible
  labels, announce validation/loading and screen changes, and preserve focus
  when expanding inline help.
- Keep headings, labels, example values, and actions sufficient to understand
  each task. Omit subtitles that repeat them. Retain concise guidance only for
  a meaningful consequence, optional help, or an actionable error.
- Required-field errors must identify the relevant field, such as `Enter a
  household name.` or `Enter an inventory name.`, rather than suggesting a
  connection or sign-in failure. Keep recoverable draft values and distinguish
  invalid addresses, connection failures, and authentication failures when the
  application can reliably do so. Never expose raw token failures, OAuth
  internals, or local fixture details outside explicit developer diagnostics.

### Invitation scope

Existing invitation parsing, browser acceptance, pending-link preservation,
preview/explicit acceptance, identity checks, and server-binding constraints in
`Invitation Deep Links` remain authoritative. Where invitation preview is shown,
identify the inventory and its authorized access/expiry details, use `You’re invited` and `Join inventory`,
and show household identity only if supplied by the authorized preview contract,
and omit copy that merely restates the join action. Show `Sign out and start over`
where the user needs to leave the signed-in invitation flow; apply the cleanup
above and clear the abandoned in-memory invitation/navigation reference.
This UX revision does not add invitation server-prefill, arbitrary-domain native
handoff, or trust in an invitation hostname as an API address. The prototype's
sample invitation journey is not implementation evidence for those capabilities.

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
- Onboarding UI/application tests must cover one-action connection/sign-in,
  address errors, browser cancellation with address preservation, returning-user
  discovery, both setup branches, name validation before writes, partial creation
  and safe retry, restart discovery, start-over cleanup, duplicate-submission
  prevention, and stale callback/result rejection. Use fakes rather than mocks.
- Before implementing changes to authentication or creation interactions, update
  adversarial end-to-end tests at the real boundaries for unauthenticated,
  expired-session, wrong-role, cross-tenant, malformed-token, and escalation
  attempts where applicable, alongside permitted-user success cases.
- Native review must verify keyboard visibility, reachable primary/ghost actions,
  VoiceOver/focus behavior, Dynamic Type, small-screen scrolling, light/dark
  appearance, and returning from canceled browser sign-in. HTML prototype review
  alone does not establish native behavior.
- Local Dex verification must prove the API accepts an ID token with the mobile
  client audience and rejects wrong-audience tokens.
- Local mobile OIDC verification must prove the mobile public client can obtain
  and refresh an ID token through authorization code with PKCE.
- `pnpm --dir apps/mobile test` and `pnpm --dir apps/mobile check` must pass.

## Invitation Deep Links

- Mobile must recognize the app-owned `stuffstash://invitations/accept` deep link and may also recognize configured HTTPS universal/app links whose path is `/invitations/accept`.
- The public mobile binary must derive its API origin from the server the user selects during onboarding. It must not embed one self-hosted API origin or tenant as production identity.
- Apple and Android verified-link declarations are static build capabilities and cannot be expanded to arbitrary self-hosted domains after installation. The general TestFlight build therefore defaults to no associated domain or verified app-link host. An operator building for one known deployment may configure those capabilities from an explicit public invitation origin.
- A configured build-time invitation origin must be an HTTPS origin on the standard HTTPS port with no credentials, path, query, or fragment for verified universal/app links. An explicit non-production local-development switch may instead trust a loopback or RFC 1918 HTTP origin for browser acceptance from a physical device; that origin must produce no iOS associated-domain or Android verified app-link declaration. A release build must fail before native generation when an explicitly configured origin is invalid, but absence of an origin is valid for the general self-host-capable build.
- Because the native iOS project is checked in, its build must materialize the same configured origin into the signed `com.apple.developer.associated-domains` entitlement; validating Expo config alone is insufficient.
- A deployment whose build enables verified HTTPS invitation links must publish an Apple app-site-association document for the built iOS application identifier and an Android Digital Asset Links document for the built Android package and signing-certificate SHA-256 fingerprint. Deployment configuration must generate these documents without requiring source edits, and the web server must serve them as JSON without redirects.
- The browser acceptance page must retain a complete browser path for arbitrary self-hosted invitation origins. A future browser-to-app custom-scheme handoff must carry a validated source server origin and must require an exact saved-server match or an explicit server switch/onboarding decision before any invitation API call; until that cross-server binding is implemented, the browser must not generate a token-bearing custom-scheme handoff. Introducing a central `stuffstash.org` universal-link broker would create a new hosted trust and privacy boundary and requires a future spec before implementation.
- The app must parse tenant, inventory, and invitation IDs from the query and the raw acceptance token from the fragment. Missing, duplicate, malformed, or oversized values must fail closed before any API call.
- Pending invitation material may remain only in memory while the app is active and in the native navigation URL needed to complete the flow. It must never enter the connection profile, AsyncStorage, SecureStore session record, diagnostics, observability, crash messages, or ordinary logs.
- A foreground invitation link captured while the asynchronous initial-link lookup is still pending must remain authoritative; the late initial-link result must not overwrite it. Likewise, preview or acceptance results from an older invitation must not replace the state for a newer captured invitation.
- If connection setup or sign-in is required, onboarding must acknowledge that an invitation is waiting and return to its preview after authentication and inventory discovery. Authentication callback handling must not replace or discard the pending invitation route.
- The invitation screen must use an application query/command and generated-client adapter behind mobile-owned ports. It must provide loading, sign-in-required, preview, email-mismatch, invalid, expired, revoked, cancelled, already-accepted, accepting, retryable-failure, and success states.
- Acceptance must require an explicit 44-point action. Success must update available tenant/inventory context and offer `Open inventory` without requiring an app restart.
- The invitation acceptance surface must honor the full platform Dynamic Type range without font-size caps. At accessibility sizes it must switch to a top-aligned, scrollable, reduced-chrome layout whose headings, metadata, and action labels wrap without clipping or hiding the primary action.
- Mobile must expose inventory sharing from a permission-gated Settings destination. Inviting must support viewer/editor selection, creation, one-time link copy, the native Share sheet, expiration context, and safe retry/error behavior in light, dark, high-contrast, VoiceOver, and accessibility Dynamic Type layouts.
- Mobile invitation tests must cover parser adversaries, deep-link routing before and after onboarding, matching and mismatched identities, every terminal invitation state, explicit acceptance, context refresh, one-time link visibility, clipboard/share behavior, and token redaction.
