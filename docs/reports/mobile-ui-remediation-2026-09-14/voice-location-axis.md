# Proposal destination selection — 24 source axes

R142, baseline9be5641a plus runner-only fixture candidate. Reviewed
VoicePlanLocationRouteScreen, VoicePlanLocationScreen, VoicePlanLocationChoices,
useParentCandidates, ParentLookupQuery, SettingsChoiceRow and the route caller.
These are source findings, not simulator passes.

Pattern decision: a searchable selection view fits a potentially large set of
hierarchical locations, contextual paths, proposed destinations and unavailable
choices. Selection commits immediately to the proposal draft; Back leaves it
unchanged. This is project judgment. Apple's searching, pickers and navigation
pages were opened September16 but returned JavaScript shells; no new Apple claim
is attributed to those unread bodies.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Choose one containing location for an editable proposed creation. Explicit current-value summary distinguishes inspection from selection. |
| Navigation | Native route title and Back; selection returns immediately. Invalid direct entry offers Back to conversation. Actual sheet-to-stack transition remains pending. |
| Selection | Root, earlier proposed creations and existing matches use shared single-choice rows with checked state. Model tests cover renamed proposed parents and opaque current IDs. |
| Modality | Production conversation retains its sheet; selection is a stack route. Shared sheet configuration is reused by the fixture. Runtime presentation/dismissal remains pending. |
| Layout | Direct ScrollView owns automatic navigation and keyboard insets. No overlay footer; row reachability beneath keyboard/header needs native evidence. |
| Adaptation | Flexible rows with contextual text; no fixed screen width. iPhone/iPad transition and window geometry are unverified. |
| Typography | Shared row text wraps; no route-specific fixed text height. Normal-size long paths need native inspection before larger text work. |
| Appearance | Shared appearance palette and Settings sections; native search and Retry. Actual contrast and dark/light rendering remain unverified. |
| Localization | Search normalizes whitespace and local case; labels remain English. Path separator, translations and RTL need broader localization acceptance. |
| Imagery | No media thumbnails. Shared checked indicator supplements accessible checked state; native symbol rendering remains pending. |
| Targets | Full-row Pressables plus native Retry/search/Back. Source dimensions do not establish native hit areas. |
| Gestures | Tap selection and explicit Back; scrolling and drag keyboard dismissal have explicit command alternatives. Interactive return must be tested natively. |
| Keyboard | Native navigation search; handled taps allow selection with keyboard visible. Clear must retain an available search field. New native journey exercises this but has not run. |
| Accessibility | Radio role, checked/disabled state, contextual labels and loading/error text. VoiceOver order and search focus return remain unverified. |
| Motion | No custom selection animation. Navigation/search transitions use native adapters; Reduce Motion requires runtime acceptance. |
| Content | Current destination stays visible even when absent from suggestions. Existing lookup intentionally returns five initial/six searched suggestions; search discovers other locations. No claim of exhaustive browsing. |
| Search | Debounce isolates results by scoped query key; prior results are hidden while a new query settles. Native search filters existing candidates and proposed names. |
| Loading | Labeled loading state until current query resolves. No global refresh spinner; pending query hides prior-query choices. |
| Recovery | Distinct error with native Retry, no-match text, disabled reasons and expired-proposal fallback. Mounted checks cover these states. |
| Editing | Selection changes only one command's parent draft; names/other edits survive. Back does not commit a selection. Actual returned proposal is included in the new native journey. |
| Privacy | Scope/plan/command/status validation blocks obsolete edits; lookup remains behind existing scoped query port. No new endpoint or authorization policy; synthetic fixture is not security evidence. |
| Notifications | Not applicable: this selector neither registers nor handles notifications. External interruption is tracked under lifecycle. |
| Media | Not applicable: no capture, playback, upload or media-selection operation. Proposal photo handling remains a separate surface. |
| Lifecycle | Visit guard rejects obsolete callbacks; scope/plan mismatch renders unavailable. Mounted current/left/wrong-plan/wrong-scope cases pass. Background, native dismissal and process restart remain open. |

## Evidence

The candidate fixture uses the real conversation workspace, real interaction
provider/controller and destination route under synthetic transport/lookup ports.
It verifies retry, no-match, native search, selection and Back through one XCTest
journey. It does not approve mutations, capture audio/photos or contact a server.
The provider spans both routes so the test observes the returned production draft.

Remote validation on paul: two fixture-isolation tests,20 focused mobile tests,
TypeScript and structural checks pass (`/tmp/voice-fixture-validation.log`). The
isolation test first failed because the required runner-only routes were absent.
This establishes fixture wiring, not a reproduced native defect. Code critic
review found no production blocker; Clear now reacquires the search field and
explicitly fails if search disappears. Swift compilation and both native journeys
remain pending on macOS runners.
