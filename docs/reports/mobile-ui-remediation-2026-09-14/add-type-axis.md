# Add item type and inventory corrections

S088, inspected at `bc55f271` with M149/M150 candidates. The shared
AssetExpirationEditor is also consumed by EditAssetSheet. Date precision itself
is covered in expiration-entry-axis.md; tag editing is covered separately.

S085 (Add kind choice) is absent by design: mobile-app-tracer-bullet.spec.md says
Add must not ask users to choose item/container. CreateAssetCommand defaults to
item; quick parent creation is a different task. S091 (Add custom fields) also has
no rendered control in Add. Settings manages definitions, and no Add value editor
is implemented here. Keep both IDs for inventory history, with axes marked N/A
for these absent controls. This does not certify custom-field management or claim
feature parity with other clients.

## Item-type review across 24 axes

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | A type classifies the item and determines expiration tracking. It is distinct from base kind. Selection belongs to the current Add draft. |
| Navigation | An inline disclosure expands choices; no separate route is pushed. Existing typed assets retain their type in Edit; initially untyped assets can select one. |
| Selection | Searchable radio rows include None and descriptive expiration support. This supports a potentially long user-defined collection in place. Native menu suitability for short collections and actual native-control fidelity still need a design/render review; source alone does not establish full platform acceptance. |
| Modality | Changing type with an existing expiration uses a native confirmation because it removes that draft date. Keep Cancel. Existing ownership checks prevent stale type acceptance; see confirmation review rather than treating every alert as covered. |
| Layout | Disclosure expands within the parent form scroll view. There is no independent footer; parent native Save and dismissal controls and keyboard avoidance govern clearance. Native long-list/date expansion remains open. |
| Adaptation | Flexible rows and wrapping text have no fixed viewport size. Wide/narrow layouts and long type names still require captures. |
| Typography | Type title uses17-point text and tracking detail uses shared palette text. Names are not explicitly truncated. Ordinary long labels need review before enlarged text. |
| Appearance | Theme text and selection checkmark are used. Radio rows are React Native controls, not proof of native menu behavior. Light/dark checks remain pending. |
| Localization | Query uses locale-aware lowercase comparison, not accent folding. Labels are English. Long/RTL names and language-aware search expectations remain part of the broader localization review. |
| Imagery | Checkmark duplicates radio checked state. No photos or provider logos occur within type choice. |
| Targets | Disclosure minimum48, choice minimum48, search minimum44. Measure hit bounds with expanded content and keyboard; declarations are not acceptance. |
| Gestures | Explicit disclosure, row selection and confirmation commands exist. Form scrolling/keyboard dismissal are parent responsibilities. |
| Keyboard | AppTextInput owns query text; it is a local search, not persisted item content. Normal-size keyboard reachability, dismissal and return focus need native checks. |
| Accessibility | Disclosure exposes current type and expanded state; rows expose radio/checked/disabled. M150 adds polite no-match feedback. Native traversal, group semantics and announcement remain unverified. |
| Motion | No picker-owned animation. Shared disclosure and native confirmation still require system-setting acceptance. |
| Content | Active custom types and None are filtered locally. Tracking support is described per row. Large collections are unvirtualized; scale requires measured evidence before acceptance. |
| Search | Query narrows visible rows without publishing draft changes. M150 replaces unexplained blank results with a no-match message; clearing restores the checked choice. |
| Loading | Parent query owns loading. M149 suppresses the misleading loading editor after an initial query failure and retains cached choices during failed refresh. |
| Recovery | Both Add and Edit offer retry for unavailable types. M149 preserves name/draft and restores the editor after recovery. Native error visibility and retry with keyboard remain open. |
| Editing | Type selection merges into existing draft; changing type clears expiration after confirmation, same-type selection is a no-op. Parent Save persists; closing search does not save. |
| Privacy | Choices come from the inventory-scoped application query. Access failure removes cached availability through the existing query boundary. UI review does not replace API authorization tests. |
| Notifications | N/A: type selection does not directly schedule notifications. Expiration persistence/reminder processing is separate. |
| Media | N/A: no camera, photo, file or audio acquisition. |
| Lifecycle | Parent draft remains the value owner. Type-change acceptance uses focused visit and asset/draft/settings identity; native dismissal/return still needs acceptance. Query state must not silently replace the item draft. |

## Evidence and remaining work

M149's Add and Edit assertions failed before correction;42 related cases then
passed remotely. The additional cached-refresh branch passed all11 Add cases.
M150's empty-search case failed before correction; all10 shared editor cases now
pass. TypeScript and structural checks pass. These do not prove keyboard geometry,
native search/control fit, large collection performance or screen-reader behavior.

No kind or custom-field UI is being added to fill an inaccurate audit inventory.
Remaining item-type design/runtime questions stay explicit rather than promoting
the surface to a native pass.
