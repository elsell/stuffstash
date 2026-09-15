# Motion-axis source review

Reviewed September15 after M61. Apple guidance: [Motion](https://developer.apple.com/design/human-interface-guidelines/motion) and [Reduced Motion evaluation criteria](https://developer.apple.com/help/app-store-connect/manage-app-accessibility/reduced-motion-evaluation-criteria/). System components may adapt automatically; custom motion still needs explicit review and runtime evaluation.

Source search across mobile UI/routes identified custom animation owners in
AppFeedback, InventoryMapScreen and VoiceConversationExchange. This search is
not a runtime pass and does not cover all behavior inside native dependencies.

- Notices: useNoticeAccessibility begins conservatively, subscribes to live changes,
  and prevents a stale initial read from overriding a later event. AppFeedback
  removes translation/spring behavior when reduced motion is enabled and keeps
  explicit dismissal/actions. Existing source tests cover live changes.
- Map: custom columns, pan/snap, breadcrumb scrolling and swipe restoration use
  reduceMotionEnabled. Its initial value is false, and its initial async read can
  overwrite a later live change. The promise has no rejection handler. This is
  a confirmed source gap requiring conservative shared preference handling.
- Voice result rail: uses animated scrolling conditionally and begins with motion
  disabled, but its async initial read can still overwrite a newer live event.
  Read failure already leaves motion disabled.

M62 implements a shared useReducedMotionPreference hook for Map, voice rails and
notices. Pending/failure stays conservative, live events supersede the initial
snapshot, and subscriptions/late reads are cleaned up. Notice screen-reader
handling remains separate. Two Map animation-request cases failed baseline;
10 focused checks pass after the fix, including read failure and later live
reenablement. TypeScript/structural checks pass; critic found no blocker. Voice
rail-specific native timing is not established by the existing entity-link test. Native
Reduce Motion entry, mid-interaction changes, gesture alternatives and dependency
animations remain unverified across the141-surface inventory. No cells are marked
passed or N/A from the source search alone.
