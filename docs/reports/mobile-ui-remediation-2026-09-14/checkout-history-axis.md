# Checkout-history review

Reviewed R008 at07ba839d with the M109 candidate, September15. Sources:
AssetCheckoutHistoryScreen/Sheet, AssetCheckoutHistoryQuery, AssetHistoryTimestamp,
AssetNativeSheetOptions, shared inventory-query/cache handling and the route.
The existing native fixture exercises records, expansion, pagination and Close;
its historical results remain in native-evidence.md. No current native pass is
inferred from source tests.

The task is to inspect one asset's checkout record without leaving its detail
context. A dismissible sheet with native Close fits that bounded task. Chronological
text belongs in a scrollable list rather than cards or a calendar. These are design
judgments grounded in Apple's [sheets](https://developer.apple.com/design/human-interface-guidelines/sheets)
and [collections](https://developer.apple.com/design/human-interface-guidelines/collections)
guidance. Local loading/recovery follows the principle of preserving useful content
while loading from Apple's [loading guidance](https://developer.apple.com/design/human-interface-guidelines/loading).

| Axis | Source conclusion and remaining acceptance |
| --- | --- |
| Task | Read-only record inspection; native title/Close and chronological rows fit. No selection editor is needed. |
| Navigation | Asset overflow opens the asset-specific route; native Close calls Back. Direct-link return behavior and native focus return remain pending. |
| Selection | N/A: rows are records, not selectable values; no selection state or commands belong on them. |
| Modality | Shared sheet options and Close support bounded inspection. Native detent, swipe and tablet behavior retain M45/M61 evidence limits. |
| Layout | Direct ScrollView uses automatic content insets. Historical M61 records navigation overlap; current bounds/overlay acceptance remains pending. |
| Adaptation | Flex record body and vertically growing notes accommodate variable content in source. Narrow/tablet reflow remains pending. |
| Typography | Roles and text styles are present; normal-size long notes/actor labels need inspection. Enlarged-text remediation is deferred per user sequence. |
| Appearance | Text, error and surface colors use appearance tokens; native commands own their style. Contrast and live appearance switching remain pending. |
| Localization | Shared Intl formatter uses device locale and timezone. English labels remain untranslated; long principals and RTL layout remain pending. M67 covers prior date correction. |
| Imagery | The record list deliberately contains no thumbnails. A decorative timeline dot accompanies an explicit status label and is not the only state cue. |
| Targets | Close and Retry/pagination use native adapters. Actual hit regions, including error/retry controls at the bottom of long history, remain pending. |
| Gestures | Sheet dismissal has explicit Close; scrolling has no custom gesture-only commands. Native swipe/scroll competition remains pending. |
| Keyboard | No text entry. External keyboard dismissal/navigation and focus return from the originating detail screen remain unverified. |
| Accessibility | Native command labels are explicit; asset subtitle is a header and page errors are alerts. Reading order, status duplication, initial-error announcement and decorative-dot traversal need assistive-technology inspection. |
| Motion | No custom screen animation; ActivityIndicator is accompanied by loading text. Native sheet transitions and Reduce Motion behavior remain pending. |
| Content | Cursor pages append compact records; failed continuation preserves earlier rows. Details are bounded to180 characters by the documented safe-summary policy. ScrollView retains all loaded rows; large-history performance is unverified. |
| Search | N/A for a search/filter control: this bounded asset-specific sheet has no specified search task. Do not add optional search solely to fill this axis. |
| Loading | History and name reads start independently; page buttons expose loading and disable while fetching. M109 corrects transient name failure hiding ready history. |
| Recovery | Empty/initial/continuation errors have local states and native retry. M109 adds separate name recovery without reloading pages. Native reachability remains pending. |
| Editing | N/A: no draft, save, mutation or undo in this read-only sheet. Asset actions and activity reversals are separate inventoried surfaces. |
| Privacy | Query keys include service/tenant/inventory/asset; access-failure cache handling clears denied data. M109 retains core denial through retries in the mounted asset owner. Controlled401/403/404 cases supplement existing history-denial tests, not real API authorization certification. |
| Notifications | No screen-owned push registration/permission control. Root notification interruption and return remain part of lifecycle verification, not N/A for the whole axis. |
| Media | N/A: no camera, files, microphone, playback or attachment controls in this task. |
| Lifecycle | Reads receive abort signals and resource keys isolate assets/scopes. Mounted denial recovery now survives retry's temporary cleared error. Cold links, remount during pending denial, background/reconnect and notification interruption remain unverified. |

M109 evidence: transient500 failure reproduced hidden records before correction;
deferred401/403/404 retries reproduced premature redisplay before the denial latch.
The final16 history/query tests plus TypeScript and structural checks pass on paul.
These prove controlled data/recovery behavior, not native layout or full lifecycle
coverage. The native fixture now starts with a controlled name-read failure; the journey
checks retained records, retry reachability and name recovery before the existing
pagination and dismissal checks. Fixture preparation, TypeScript and structural
checks pass remotely. Execution on iPhone/iPad remains pending; this addition is
not new runtime acceptance evidence.

## Current route follow-up

Atc08aa31a, the six remaining R008 source axes were rechecked: automatic scroll
insets and flow footer; responsive wrapping record content; no keyboard input;
native Close, named loading/error/retry and section heading; no notification
handler in this route. Actual geometry, assistive traversal and direct-link return
remain native acceptance. M212 identifies Close calling Back even without a stack.
The candidate replaces with Home only on direct entry. Loading/error/ready tests
failed before correction and now cover both entry conditions. Nine route tests,
TypeScript and structural checks pass on paul; critic found no blocker.

September16 Android normal-text follow-up now verifies independent name recovery,
older-page append/exhaustion, retained notes, ordinary Close and cold-root Close.
See android-header-sheets.md for exact APK and retained evidence. Production detail
focus, failed continuation and assistive reading remain separate acceptance work.
