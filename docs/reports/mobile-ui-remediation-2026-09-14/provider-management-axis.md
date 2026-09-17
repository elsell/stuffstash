# Provider management — all 24 axes

Initial source checkpoint: `5501152d`; result-handling follow-up at `d39e4233`.
R052 is provider detail, R054 creates a recommended
profile, and R055 lists profiles. Reviewed their route wrappers,
`ProviderProfileScreens`, `ProviderSettingsSupport`, `VoiceAdminGuard`, and the
shared settings rows/styles. Credential and prompt editing remain covered in
[their own review](provider-editors-axis.md). This is source and mounted behavior
evidence; provider-management native captures have not been accepted.

The list opens substantial configuration resources. Creation issues a command
that makes a disabled draft, then replaces the creation route with its detail.
Detail mixes navigation into editors with explicit server commands. M148 now uses
disclosure rows for destinations and the existing native command adapter for
mutations. Archive has a native confirmation naming the affected profile. These
are project applications of the platform interaction standard, not a rule that
every settings action needs a separate screen.

## Review

| Axis | Evidence, decision and remaining acceptance |
| --- | --- |
| Task | List/detail/create serve distinct tasks. M148 separates commands from destinations. Creation explains the disabled draft and subsequent credential/test/enable steps. Verify the recommended-service commands at normal text size. |
| Navigation | Routes use the native stack. List opens encoded profile IDs; Add replaces itself with detail after creation, so Back returns to the list. Detail opens credential/prompt editors. Mounted tests cover command-related navigation suppression; native Back and deep-link return remain open. |
| Selection | No small flat preference is pushed from these three surfaces. Choosing a recommended service creates a resource; lifecycle buttons mutate it. Keep their action labels. Provider selection for voice stages is a separate surface. |
| Modality | These are pushed destinations. Archive alone requires a native risk confirmation. M146 consumes acceptance once and requires a new confirmation for retry. Native alert placement/dismissal remains open. |
| Layout | Main content is a ScrollView with bottom padding. Commands are in content, not a separate overlay footer. Shared groups clip overflowing children; that makes native command measurement and scroll-to-last-command important acceptance checks. No source claim of safe-area clearance. |
| Adaptation | Flexible content has no fixed screen height. Settings rows adapt to font scale, but these screens have no dedicated tablet column constraint. Review normal-size phone, iPad portrait/landscape and narrow windows before proposing a layout change. |
| Typography | Profile names can wrap and detail uses a prominent heading. Settings styles define nominal fonts and fixed line heights for secondary copy. Long names and normal-size hierarchy need captures; enlarged-text remediation follows the user's sequencing. |
| Appearance | Shared semantic palette supplies background, text and borders. M148 removed custom command rows in favor of NativeCommandButton. Native dark/light command contrast, disabled state and group clipping remain unverified. |
| Localization | English labels and template descriptions remain part of the existing localization gap. Provider names are external content; list sorting and last-test presentation come from shared presenters/query. Verify long names and RTL; no locale certification here. |
| Imagery | No provider logos or content photos are required. Disclosure comes from the shared navigation row. The information is textual and does not depend on missing-image fallbacks. |
| Targets | Navigation rows declare 52-point minimum height and 44-point minimum width. Commands now use the native adapter. These declarations do not establish rendered targets or reachability. Measure each command and the last scrolled action. |
| Gestures | Scrolling and native stack return provide standard interaction. Commands have explicit tap alternatives; no custom essential gesture exists. Verify back swipe during a pending command and subsequent return. |
| Keyboard | N/A within these three screens: no text entry. Credential/prompt editors have separate keyboard coverage. A keyboard left over during route transitions remains a navigation acceptance case. |
| Accessibility | Rows identify destination, profile and lifecycle; values have combined labels. Commands expose their visible names and pending labels. Native traversal, disabled announcements, grouping and focus after creation remain open. |
| Motion | No custom animation is defined by these screens. Native navigation, progress and shared feedback still need reduced-motion acceptance; do not add screen-specific animation machinery without evidence. |
| Content | List shows profile name, stage/provider and lifecycle. Detail exposes model, status, credential status and last test without secret values. Empty list offers Add Profile. All returned profiles are rendered; large collections are a scale question requiring evidence, not justification for unsolicited search. |
| Search | N/A for the current setup task: no search/filter requirement is established for the profile collection. Reassess with observed collection size and findability. This does not certify unbounded-list performance. |
| Loading | M180 labels initial loading by task; transient refresh retains data with SettingsRefreshNotice. Creation and detail commands use synchronous locks plus operation-specific pending labels. A blocked command cannot start a second operation. Native pending-state presentation remains open. |
| Recovery | M180 corrects the shared Voice Setup failure heading to the current task. Query failure offers Retry; missing detail names the unavailable profile condition; empty list remains actionable. M181 rejects fulfilled failed test results instead of falsely showing Connection tested. Mounted cases cover failure and successful retry on both detail and stage settings. Native feedback presentation and focus remain unverified. |
| Editing | These screens contain no text draft. Creation first stores a disabled profile; the editor owns later secrets. Archive is a deliberate mutation with confirmation. M146 protects single-use acceptance. Partial setup remains accessible from the profile list. |
| Privacy | Route wrappers require configure permission through VoiceAdminGuard. Query access failure suppresses cached profiles. Displayed credential status is not the secret. UI inspection does not verify server authorization; backend boundary evidence is separate. |
| Notifications | N/A for OS notification entry/scheduling. In-app feedback belongs to recovery and lifecycle. |
| Media | N/A: these configure voice providers but do not record audio or acquire photos/files. |
| Lifecycle | Creation/detail completion is scoped to command/resource and focused visit with useTaskPresentation. Existing mounted cases cover late completions, pending exclusion and confirmation reuse. Native blur/return, backgrounding and cold detail entry remain open. |

## Acceptance still required

Run list empty/loaded/error/refresh, create success/failure, unavailable detail,
test/enable/disable, and Archive cancel/confirm/failure/retry. Inspect normal-size
phone and iPad, long profile names, both appearances, last-action scrolling and
navigation return during pending work. Capture visible controls and hit bounds.
Then verify assistive traversal and adaptation in the later accessibility pass.

The 57 settings mounted cases passed remotely for M148, along with TypeScript and
the structural check; critic found no confirmed regression. Those results preserve
behavior evidence, not native acceptance. M146/M148 remain finding cells until the
relevant acceptance evidence is recorded.
