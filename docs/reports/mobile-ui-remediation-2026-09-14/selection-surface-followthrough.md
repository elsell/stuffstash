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

| Adaptation | Selection routes use fullScreenModal on iOS and card navigation on Android. Shared rows are flexible; they have no fixed screen width. Tablet/window composition is not established by source. |
| Typography | RN text keeps default font scaling, labels can wrap, and SettingsList stacks row contents at fontScale1.3. SwiftUI footer labels use vertical fixedSize for wrapping. No maximum text size is imposed here; enlarged-text runtime remains open. |
| Localization | Labels are English literals; tag ordering/search use locale-aware string operations. The views have no date/time formatting. RTL visual order and translations are not established; this source review does not imply localization support. |
| Imagery | Selection state uses the shared20pt checkmark and explicit checked semantics, without photos or color-only distinctions. System header/search imagery is delegated to existing native adapters. |
| Targets | SettingsChoiceRow specifies minimum52pt height/44pt width; the segmented adapter specifies44pt height. iOS footer uses large native bordered controls and Android uses Compose Button/OutlinedButton. These are configured dimensions, not measured delivered touch regions. |
| Gestures | Explicit Cancel is available in both tasks. Destination route removal invokes owned cancellation and respects blocked work; selection uses platform presentation. Both scroll bodies support drag keyboard dismissal. Native gesture/back races still require scoped evidence. |
| Accessibility | Choice rows expose checkbox/radio with checked/disabled state and explicit labels. Footer commands have explicit native labels; headers expose heading roles and tag count uses a live region. Screen-reader traversal and announcement quality require VoiceOver/TalkBack verification. |
| Motion | These consumers introduce no custom animations; transitions and search use platform adapters. This avoids a new custom-motion owner but does not certify Reduce Motion behavior of the delivered presentation. |

Sources: `AssetTagSelectionScreen`, `AddDestinationSelectionScreen`, their route
screens, `AssetTagSelectionTask`, `AddDestinationTask`, both NativeFilterSheet
adapters, `AssetNativeSheetOptions`, `SettingsScreen.styles`,
`SettingsScreenPresentation`, `NativeSegmentedControl` and native NativeSheetActions.
All previously unreviewed axes now have source review; none is promoted to runtime
acceptance by this report. The later iPad Add creation/retry refresh36100666455 passes; see README
for the scoped phone observation failure and remaining physical/assistive gaps. No new product correction is justified by this source pass.
