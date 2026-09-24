# Selection surface follow-through

Source review at0610209a. This records implementation ownership, not native
acceptance. The current route inventory covers every `.tsx` route; none of its
primary source paths are missing. No additional route was silently omitted.

The two recently added selection routes use dedicated task presentations because
locations need search and contextual paths, while tags need searchable multiple
selection and a staged Done/Cancel decision. The review below addresses their
previously unreviewed source axes without treating each matrix cell as a test.

| Axis | Source evidence |
| --- | --- |
| Layout | AddDestinationSelectionScreen owns an automatically inset ScrollView, iOS keyboard adjustment and Android KeyboardAvoidingView/bottom safe area. AssetTagSelectionScreen uses NativeFilterSheet: measured opaque footer with reserved body space on iOS; separate flex body/footer on Android. These sources explain ownership but cannot prove rendered clearance. |
| Appearance | Both use the appearance palette and SettingsSection/SettingsChoiceRow; selection checkmarks use the palette action color. No fixed light-only color in these consumers. Native dark/contrast acceptance remains separate. |
| Content | Destination puts current location/top level before suggestions and gives creation its own state. Tags separates All/Selected, sorts labels, reports selection count and retains removable unavailable selections. Long-list performance remains unverified. |
| Loading | Destination exposes loading, failed suggestions/retry and creation progress while guarding selection/creation during busy states. Tags receives its options from the owning editor and explicitly handles unavailable selection and empty results. Owner loading coverage is outside this local source conclusion. |
| Privacy | These presentation routes receive editor-owned data; they do not fetch arbitrary inventory IDs. Tag completion checks the original scope and owner availability. This is presentation ownership evidence, not API authorization certification. |
| Lifecycle | Providers assign each visit an owner; clearing an old owner cannot erase another visit. Tags rejects obsolete visits/scope changes. Destination cancellation checks the active visit and blocked state; route removal cancels the task and routes with no task exit. Native background and gesture races remain scoped to existing evidence, not newly certified here. |

Sources: `AssetTagSelectionScreen`, `AddDestinationSelectionScreen`, their route
screens, `AssetTagSelectionTask`, `AddDestinationTask`, and both NativeFilterSheet
adapters. Typography, window adaptation, localization, imagery, touch-region
geometry, gestures, screen-reader order and motion remain explicitly unreviewed
in this follow-through. Existing iPad Add readiness and physical/assistive gaps
remain open in README; no new runtime claim or product fix is made.
