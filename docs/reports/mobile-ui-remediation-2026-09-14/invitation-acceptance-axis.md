# Invitation acceptance review

R019, source review at6f45c031, normal text first. Entrypoint:
`app/invitations/accept.tsx` → `InventoryInvitationScreen`; system-link intake
S131 is related but is not certified by this route review. The route delegates
opening/dismissal to `useInvitationRouteActions`. Existing mounted progress and
onboarding tests cover explicit acceptance, pending labels, opening recovery and
replacement-reference failures. The older behavior suite mocks React hooks and
is weaker evidence for actual effect ordering.

The task is to review inventory/access/expiry, explicitly join, then open the
inventory. A dedicated review screen fits this consequential external-link task;
a small selection menu would not replace it. M133 concerns command implementation,
not the need for this screen. Apple's Buttons and Progress indicators topic URLs
were consulted September15; the web fetch returned JavaScript-only shells, so this
report does not attribute specific wording to those pages. The native-adapter
requirement below follows the repository's established platform interaction policy.

| Axis | Source assessment and remaining acceptance |
| --- | --- |
| Task | Review → explicit Join → Open is appropriate; M133 custom command family remains. |
| Navigation | Successful route action clears link and replaces Home only for its focused owner. Actual back/deep-link return needs runtime. |
| Selection | Viewer/editor is displayed as granted access, not an editable choice. No selection control required here. |
| Modality | Full route contains review and terminal states. Not now/Done dismiss through route actions; swipe/system Back behavior needs runtime. |
| Layout | Flexible ScrollView, centered max-width card, no absolute action footer. Native insets and last-action clearance unverified. |
| Adaptation | Card maxWidth520 and narrow-width branch exist. Phone/iPad geometry unverified. |
| Typography | Labels wrap without line limits; enlarged-text branch exists but that pass is deferred. Normal long-name fit remains unverified. |
| Appearance | Uses theme tokens; custom command disabled styling is not a native disabled treatment (M133). Contrast not measured. |
| Localization | Expiry uses Intl default locale and medium date/short time. English text and RTL rendering need verification; no localization completion claim. |
| Imagery | Brand mark and success/invitation symbols supplement text; no photos involved. Icon accessibility traversal unverified. |
| Targets | Custom controls declare54-point minimum height; actual native targets pending M133 migration and runtime. |
| Gestures | Commands are explicit. System Back and retained pending operation behavior need runtime. |
| Keyboard | This screen has no text input; verify keyboard dismissal when entered from sign-in. |
| Accessibility | Live region wraps card and Join/Open declare busy/disabled. Native reading order, duplicate announcements and focus transitions unverified. |
| Motion | Loading and operation indicators accompany named states. Reduce Motion behavior unverified. |
| Content | Shows inventory name, access and expiry without pagination. Expired/revoked/cancelled states explain next step. |
| Search | No search task in invitation review; N/A. |
| Loading | Loading/Joining/Opening are named; pending acceptance disables Join and start-over. Start-over has no separate visible pending wording; review in command migration. |
| Recovery | Invalid/nonpending invitations do not offer acceptance; transient preview errors retry, mismatch offers account switch, opening failure preserves accepted access. Mounted opening recovery exists. |
| Editing | No editable draft. Not now does not undo accepted access. Repeated command dispatch and interruption still require runtime/mounted review. |
| Privacy | Preview/accept use injected application ports; error classes separate auth/mismatch/invalid response. This source review does not certify authorization or link validation. |
| Notifications | No notification-specific entry handling in this route; external invitation links tracked separately in S131. |
| Media | No acquisition/playback/upload in this route; N/A. |
| Lifecycle | Request generation rejects obsolete reference responses; mounted replacement-reference cases exist. Focus return/hidden callbacks/cold-warm real links remain unverified. |

M133 acceptance: use the existing native command adapter for all commands; preserve
Join/Open pending feedback and disabled semantics, explicit acceptance, account
switch, start-over recovery and accepted-access explanation after failed opening.
Verify normal phone and iPad entry, long inventory name, pending states, failure,
retry and dismissal. Do not infer runtime acceptance from adapter tests.
