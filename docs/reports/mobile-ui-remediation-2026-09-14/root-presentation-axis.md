# Root presentation and service gate

R006 source review at bcaeadfa with M152 ownership candidate. Inspected root
_layout, AppServicesContext/FeedbackGate/Gate, AppearanceContext,
AppKeyboardProvider.ios, AssetNativeSheetOptions, InventoryInvitationLinkContext,
PushNotificationNavigation and existing related audit appendices.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Root composes authenticated navigation, onboarding and global presentation. It is infrastructure for tasks, not another user destination. |
| Navigation | Native root Stack contains tabs, settings and asset tasks. Nested tab stacks keep peer navigation separate. Actual Back/return and deep-link stack composition need native evidence. |
| Selection | N/A: no root-owned value picker. Inventory selection and sheet controls have separate surface IDs. |
| Modality | Add is a full-detent form sheet; filters and other tasks use named sheet options; voice has two detents. These are individual task choices, not a universal modal rule. Child dismissal guards remain necessary. |
| Layout | AppNoticeScreenLayout owns ready-session notices per screen. Keyboard accessory is mounted beside the root stack; M103/M137 track the actual layering/hit-testing gaps. Root declarations do not establish safe-area acceptance. |
| Adaptation | Platform native stacks/sheets govern window adaptation; several detents and corner radii are explicit project choices. iPad/Android/compact layout remains incomplete. |
| Typography | Native header text uses palette and explicit bold weight. Route titles/Back are English. Long ordinary titles and toolbar crowding require runtime inspection. |
| Appearance | Appearance hydration precedes service composition; status bar follows resolved scheme. Preference/contrast read failure falls back rather than permanently withholding content. Cross-surface theme changes remain in appearance audit. |
| Localization | Root title/Back literals are not localized. No blanket localized or RTL acceptance; see localization-axis.md. |
| Imagery | Root owns no product images. Native back/route affordances and child icons remain separate. |
| Targets | Native headers/sheets own root touch geometry. Shared keyboard accessory is an exception requiring actual bounds, tracked separately. |
| Gestures | Native back and sheet gestures vary by route. Edit/return tasks intentionally disable gesture dismissal; child explicit cancellation must remain reachable. |
| Keyboard | iOS KeyboardProvider disables preload; other-platform wrapper is a fragment. Global accessory and per-screen text ownership remain distinct. Existing native keyboard failures are unresolved, not closed by this review. |
| Accessibility | Status bar and native header semantics are delegated. Loading screen has text and an ActivityIndicator. Initial focus, announcement, root-to-sheet traversal and accessory ordering need runtime acceptance. |
| Motion | Native stack/sheet transitions are delegated. Reduced Motion behavior is not verified by root composition. |
| Content | Route registry is finite; no root data list or pagination. Children retain their own data-state audits. |
| Search | N/A: root owns no search state; Browse, filters and detail own scoped search. |
| Loading | Appearance hydration renders a blank themed view, then service gate renders Loading Stuff Stash. Startup profile-resolution error returns instance onboarding. Delayed storage and perceived launch progress remain unverified. |
| Recovery | Gate preserves current session if push cleanup rejects sign-out/change-server. Startup error has no dedicated retry explanation; native initial-storage failure recovery remains an open acceptance scenario. |
| Editing | Root has no asset draft. Shared providers and task routes must preserve child work; navigation options alone do not prove that. |
| Privacy | Service context mounts only when ready; server-state provider uses the composition scope. M152 prevents old-composition expiry callbacks from expiring a replacement session. This does not claim full authorization coverage or cancellation of already-started credential operations. |
| Notifications | Push navigation waits for a root key, aborts previous open attempts, unsubscribes on teardown and routes resolved asset IDs. Physical cold/warm push acceptance remains separate. |
| Media | Root mounts voice state/return controls; acquisition and permission belong to child tasks. Audio background and camera/library interruption are unverified here. |
| Lifecycle | M152 guards expiry initiation/completion by composition identity and retires the owner on successful onboarding transitions or root cleanup. Expiry dialog dismissal has no session mutation. Invitation capture starts outside the authenticated gate. Account switching, process restore and native notification interruption still require end-to-end acceptance. |

Evidence: source inspection and mounted gate/onboarding tests. No new screenshot,
physical-device result or full-root runtime acceptance is claimed. Related
appendices: tab-shell-axis.md, global-notice-axis.md, keyboard-accessory-axis.md,
appearance-axis.md, invitation-acceptance-axis.md and notifications-axis.md.
