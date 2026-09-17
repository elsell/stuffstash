# Browser sign-in return — 24-axis source review

S059 atc7b64036 plus M202. Reviewed OnboardingScreen/Presentation,
OnboardingCommand and both ExpoOidcNativeClient adapters. System browser sign-in
keeps provider credentials outside the inventory form. This review distinguishes
the app's return handling from the provider's web content and the OS session UI.

| Axes | Evidence and acceptance limits |
| --- | --- |
| Task, navigation | Connect and sign in opens authorization, then resolves existing authorized inventories or required household setup. Successful return alone does not mean onboarding is complete. No extra app sign-in selection screen is introduced. |
| Selection, editing | Server address remains editable before/retry after the attempt. While submitting, pending guard and disabled controls prevent competing submissions. Password/account choices belong to the provider/system session, not app-owned fields. |
| Modality, gestures | Expo AuthSession/WebBrowser owns browser presentation and callback. Explicit cancel/dismiss returns to retry without a created session. Real swipe/dismiss and return-to-app behavior remain physical/native checks. |
| Layout, adaptation, typography | App form is scrollable, width-bounded600 and has automatic keyboard insets; headings/body wrap. Browser chrome and provider content are external. Existing onboarding fixture passes cover form/keyboard behavior, not live OIDC return. |
| Appearance, imagery | App semantic palette and text-first instruction; no image is needed to interpret return. Provider branding, light/dark browser chrome and contrast are outside the source verification. |
| Localization | English app guidance, URL normalization and clear connection label. Provider language, RTL and long translated errors remain unverified. |
| Targets, accessibility | App controls have labels/disabled state, and title changes request heading focus. Verify actual browser-return focus, error announcement and reachability; props do not establish screen-reader behavior. |
| Keyboard | URL keyboard disables capitalization/correction. System browser owns its credential keyboard/autofill. Return with app keyboard retained/dismissed requires a live native session. |
| Motion | No custom auth animation. System browser transitions and app heading focus need Reduce Motion/assistive checks. |
| Content, search | Brief explanation says the browser opens and returns. Contextual connection help remains available. No app-owned result list, pagination or search in this task. |
| Loading, recovery | App shows pending submission and keeps draft on failure. M202 distinguishes cancel/dismiss from error/locked/unknown results. Provider parameters never enter the new message. Retry succeeds through the composed boundary. |
| Privacy | Authorization-code PKCE, returned state equality and usable code/verifier are required before token exchange. Composed checks reject malformed returns without session or tenant discovery. These fakes do not certify the OS callback association, provider behavior or server token verification. |
| Notifications, media | No push permission or camera/microphone request in this task. A pending invitation can survive onboarding through the separately audited root owner. External interruption remains a native scenario. |
| Lifecycle | OnboardingCommand serializes work and checks generation after awaits; reset/dispose supersedes it. Screen completion checks its mounted generation. Browser backgrounding is expected, not a reason to invalidate an otherwise valid authentication response. Process-death restoration and live callback handoff remain unverified. |

M202 is feedback correction only. Actual provider/session errors previously threw
the same cancellation error as explicit cancel. Three failing composed-boundary
cases reproduced that message before the fix; successful retry is now checked
for each. Cancel and dismiss retain their existing semantics.

All52 related checks across5 files, TypeScript and structural checks pass remotely
on paul (`/tmp/onboarding-return-green.log`). This is controlled boundary/source
evidence, not a successful real browser login or physical-device acceptance.

Critic found no production blocker and requested direct exchange-count evidence.
The composed cases now assert zero exchanges after failure and one after retry.
All7 boundary cases and static checks pass again (`/tmp/onboarding-return-reviewed.log`).
