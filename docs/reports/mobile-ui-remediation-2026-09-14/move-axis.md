# Move surfaces: source review

Scope: S134 destination search/selection, S135 destination creation, S136 Move
here search/selection. Source at3df1d610. This is a source review, not native
acceptance. Follow the platform-interaction decision framework; large/searchable
containment choices justify a selection view rather than a small value menu.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Move changes one asset's parent; Move here chooses an asset for the current destination. Separate tasks have explicit previews. M91: command footers remain custom despite the native adapter. |
| Navigation | Native form-sheet routes return through router.back after success. Actual entry/return and scroll restoration remain unverified. |
| Selection | ParentRow exposes selected/disabled states and a visible Selected label. Move excludes self and invalid destinations; Move here excludes current children. Native reading order and row reachability remain pending. |
| Modality | Bounded tasks use sheets. Move/Move here retain dismissal while idle and guard busy dismissal. Large-text detents and cancellation need runtime review. |
| Layout | Titles, help, query and placement preview occupy fixed space above results. Move here also has a fixed bottom preview. These can crowd results at large text; do not infer resolution from Edit's scroll fix. |
| Adaptation | Flexible widths exist, but fixed regions and horizontal preview/row composition need phone, iPad and large-text evidence. |
| Typography | Text permits scaling; explicit font sizes/line heights are not proof that large labels fit. Long titles, paths and buttons require capture. |
| Appearance | Styles use the appearance palette. Contrast, button differentiation and accessibility appearance modes remain unmeasured. |
| Localization | User-facing strings and preview arrow need expanded-string/RTL review. Path order and directional meaning are not established by source tests. |
| Imagery | These selection rows use title, kind and path rather than photos. Confirm this is sufficient for similarly named assets in realistic inventories. |
| Targets | Rows and footer expose buttons; native target bounds and crowded touch regions remain unverified. |
| Gestures | Selection and completion have explicit controls. Sheet drag conflicts with result scrolling/keyboard still need native execution. |
| Keyboard | AppTextInput supplies shared entry behavior; result scroll adjusts keyboard insets inside KeyboardAvoidingView. Actual focus, dismissal and hardware keyboard navigation remain pending. |
| Accessibility | ParentRow selected/disabled semantics are present. Full VoiceOver/TalkBack traversal, announcements after retry/selection and focus return remain pending. |
| Motion | No additional animation in these form implementations. Native sheet motion and shared keyboard animation still need Reduce Motion checks. |
| Content | Rows show title, kind and path; From/To or Move preview communicates the proposed change. Very long paths and dense results remain pending. |
| Search | useParentCandidates debounces250ms, uses a query-specific cache key and hides data until the query settles. M90 separates unknown results from empty results in Move here. Native typing and clear/return behavior remain pending. |
| Loading | M90 fixes false empty results in Move here. Move keeps CandidateStatus above its entire form; unknown suggestions also feed destination-creation eligibility as an empty array, requiring a deliberate product decision. |
| Recovery | M90 retry lives in Move-here results and preserves query; existing operation tests retain drafts after failures. Native alerts, retry reachability and selection retention remain pending. |
| Editing | Commands pass through application objects and synchronous busy guards. Existing tests cover draft freezing/failure recovery. Creation followed by cancellation and duplicate destination behavior need end-to-end review. |
| Privacy | Asset permission gates and scoped queries are used. This source pass is not adversarial authorization verification. No permission changes made. |
| Notifications | No direct notification controls in these forms. App-level notification interruption and return still apply; do not mark all lifecycle effects N/A. |
| Media | No acquisition or upload controls here. No media-specific pass claimed for the containing app. |
| Lifecycle | Operation ownership suppresses stale completion after teardown/replacement. Backgrounding, inventory switching and physical-device interruption remain unverified. |

M91 applies to shared SheetActions in AssetDetailSheets.tsx, used by Edit,
Move and Move here. It renders custom Pressable/brand-styled commands. Existing
NativeSheetActions and NativeCommandButton adapters establish that native commands
are available, but substitution must preserve both primary and secondary busy
semantics and avoid adding a second keyboard-avoidance owner. Native footer
measurement issues remain a verification concern, not a reason to silently retain
an undocumented custom control indefinitely.

Next: native recovery/selection journeys; resolve the fixed-content layout and
footer adapter choice; explicitly decide destination creation when lookup fails.
