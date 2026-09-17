# Inherited settings detail

S109, source review at438bd902 with M153/M154 candidates. Inspected
CustomizationEditorScreen, CustomizationEditorFields, CustomizationRoutes,
CustomizationCollectionScreen, SettingsList and mounted customization tests.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Inspect a shared definition and identify where it can be managed. Read-only detail is appropriate; it must not imply a local override. |
| Navigation | Collection groups identify inherited records. Manage in household replaces the detail route with its household editor, retaining the record/lifecycle. Actual Back destination needs native review. |
| Selection | No editable choice for inherited records. Loaded ownership overrides route hints. Field type, applicability and persisted options use static values. |
| Modality | Detail is a native stack route, not another nested sheet. Inherited inspection has no draft-discard confirmation. |
| Layout | Shared settings inset and parent scroll content contain values, Details and management action. Long values and final-action clearance remain native checks. |
| Adaptation | Flexible text has no fixed viewport dimensions. iPad and narrow-phone presentation are unverified. |
| Typography | Labels and values wrap. Ordinary long household/definition names precede enlarged-text checks. |
| Appearance | Theme palette supplies text and grouped surfaces. M154 removes an inactive switch from read-only asset-type detail; measured contrast remains open. |
| Localization | English inheritance/management text includes user-provided household names. Long/RTL names and translation remain separate pending acceptance. |
| Imagery | Details disclosure uses a chevron; this surface has no photos. Decorative icon behavior and native disclosure rendering remain unverified. |
| Targets | Details and Manage are explicit actions. Shared minimum row geometry is source evidence, not measured native hit bounds. |
| Gestures | Explicit Details/Manage and native Back provide alternatives to navigation gestures. |
| Keyboard | No editable inputs should appear in inherited detail. Keyboard/focus after returning from the household editor remains a runtime check. |
| Accessibility | Scope is exposed as a labeled value. M154 makes expiration tracking text instead of a disabled switch. Group reading order and focus on route replacement remain unverified. |
| Motion | No surface-owned animation beyond disclosure state and navigation. System Reduced Motion behavior remains pending. |
| Content | Loaded field/type values populate the detail; unavailable persisted targets retain a count. M153 corrects the Scope label to the actual owner. |
| Search | N/A: detail has no search. Collection search and editable target selection are separate surfaces. |
| Loading | Context/definition loading precedes detail; stale resource loads have workflow ownership. Query failures render recovery instead of an empty editable form. |
| Recovery | Denial removes access; transient definition refresh failure has explicit retry. Current data/error visibility and native retry reachability remain open. |
| Editing | No Save or lifecycle mutations for inherited records. The separately authorized household editor owns changes; M154 displays Enabled/Disabled for tracking. |
| Privacy | Actual loaded record scope, not inherited route hint, governs read-only ownership. Household management action additionally requires configure permission. Existing policy/boundary tests remain required. |
| Notifications | N/A: inspecting a definition does not schedule notifications. Changing tracking in the owning editor is a separate task. |
| Media | N/A: no camera, library, files or audio controls. |
| Lifecycle | Resource changes invalidate pending loads. Navigation return, background refresh and permission changes during native inspection need runtime acceptance. |

Both inherited-owner cases and both tracking-value cases failed before correction.
All49 customization behavior tests, TypeScript and structural checks now pass
remotely. M153/M154 are source/mounted corrections; no native screenshot acceptance
is claimed. Existing tests emit act warnings in other customization cases; those
warnings are not evidence that native navigation or asynchronous layout passed.
