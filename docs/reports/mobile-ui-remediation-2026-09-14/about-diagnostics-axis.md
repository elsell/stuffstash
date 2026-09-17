# About and Diagnostics — 24-axis source review

R023/R027, source997bfa50 plus M183. Inspected route wrappers, root stack entries,
SettingsDetailScreens, SettingsList, SettingsQuery and scoped query hooks.
These are read-only destinations for app identity and troubleshooting. They need
neither choice menus nor a new modal. Source review does not certify native fit.

| Axis | Evidence and remaining verification |
| --- | --- |
| Task | About explains the app and version. Diagnostics deliberately exposes technical connection/identity values separately from everyday settings. |
| Navigation | Both are named native stack routes. Native Back and direct-entry return remain unverified. |
| Selection | N/A: read-only values, no preference choices. |
| Modality | No sheet or confirmation; native pushed destinations fit supporting information. |
| Layout | ScrollView and shared grouped sections, with content padding rather than overlay commands. Inspect long endpoint/ID wrapping and bottom clearance. |
| Adaptation | Flexible shared rows support wrapping; tablet widths and landscape need native captures. |
| Typography | Shared settings typography; About has a heading/subtitle. No large-text acceptance claim. |
| Appearance | Shared semantic palette; native Retry commands. Light/dark contrast and pending rendering remain open. |
| Localization | English copy; long server URLs and identity values are data. RTL/layout remains unverified. |
| Imagery | N/A: no images or essential symbols. |
| Targets | About has no body command. Diagnostics uses native Retry; actual hit geometry remains open. |
| Gestures | Standard scrolling and native Back; no custom essential gestures. |
| Keyboard | N/A: neither screen accepts text. |
| Accessibility | Shared value labels and section headings; named loading rows and error alerts. Verify traversal, announcements and focus after recovery. |
| Motion | No custom animation. Native transitions and loading indicators need reduced-motion review. |
| Content | About keeps version local. Diagnostics groups connection, identity and application. M183 preserves local information when remote reads fail, without blank identity placeholders. |
| Search | N/A: small fixed set of values; no list discovery task. |
| Loading | About needs no remote reads. M183 gives account and household identity independent loading states; local diagnostics never wait for either. |
| Recovery | M183 gives each failed identity query a native Retry and safe copy, retaining stale non-access-failure values with disclosure. Repeated retry is disabled while fetching. |
| Editing | N/A: no drafts or mutations. Retry only reads the relevant identity. |
| Privacy | Existing scoped hooks and keys retained. Access failures hide cached principal/tenant IDs; arbitrary transport messages are not displayed. The mounted denial case is UI evidence, not a replacement for backend security tests. |
| Notifications | N/A: no scheduling, permissions or OS notification entry. |
| Media | N/A: no capture, file selection or playback. |
| Lifecycle | Local diagnostics read the current injected provider each render. Queries use session/inventory keys and cancellation; no completion navigation. Native cold start, backgrounding and return remain open. |

M183 is a recovery/task-fit correction under the project interaction standard,
not an assertion that Apple prescribes these exact sections or labels. Two mounted
RED cases reproduced hidden local diagnostics during pending/failed discovery and
after access loss. After the change,68 settings cases plus TypeScript and structural
checks pass on paul (`/tmp/diagnostics-reviewed.log`). The expanded denial case
also covers a subsequent transport failure and successful retry. The existing
production QueryCache clears denied data, so no new suppression latch is needed.
No native result is claimed.

Next native acceptance: open both pages, inspect long configured values, then
pending identity, failed identity, successful retry, stale refresh and access loss.
Verify local connection/version remain available throughout. Include normal-size
phone/iPad and both appearances before enlarged-text follow-up.
