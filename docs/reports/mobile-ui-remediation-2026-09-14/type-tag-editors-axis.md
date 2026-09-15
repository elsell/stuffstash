# Type and tag editors — all 24 source axes

R028/R030/R036/R038/R045/R047 at9e61709c plus M165. Reviewed shared
CustomizationEditorScreen, route composition, workflow/permission use, native Save
adapter and lifecycle section. Field-specific controls are not certified by this
type/tag review, although their shared Save consumer receives M165 too.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Create/edit named tags or asset types; types optionally track expiration and have a description. Read-only/inherited values use static text. |
| Navigation | Route completion replaces with its collection; inherited management replaces with household editor. Back currently uses a custom leading control rather than native bar item: M166 remains open. |
| Selection | Type expiration uses shared native switch; tag color uses system color adapter. These are local draft choices, not separate stack tasks. Color activation/target failures remain M51. |
| Modality | Dirty exit uses system confirmation, owned by focused resource. Lifecycle mutation also confirms before acting. Native dismissal and confirmation interaction remain acceptance work. |
| Layout | Scroll content uses automatic insets and keyboard adjustment inside KeyboardAvoidingView. M165 keeps Save in the shared content column. Actual keyboard inset interaction, safe areas and bottom reachability remain unverified. |
| Adaptation | Single-column grouped form. Normal-size long names/descriptions and tablet windows need native captures; enlarged-text work stays queued. |
| Typography | Labels, required-name error, scope and read-only values share settings styles. Description is multiline; key details are progressively disclosed. Native wrapping/legibility remains pending. |
| Appearance | Semantic palette and native primary Save after M165. Remaining custom Back/lifecycle command controls are tracked under M166, not treated as native merely because they resemble settings. |
| Localization | User names/descriptions remain verbatim; labels are English. Auto-generated stable keys can require manual correction, revealed/focused on invalid key. Non-Latin input and RTL runtime acceptance remain open. |
| Imagery | No photos. Color is optional tag metadata with a named read-only fallback; it is not the only identifier. Disclosure icons supplement text. |
| Targets | Native Save inherits shared adapter sizing; custom Back/disclosure/actions declare minimum dimensions. Actual hit regions remain runtime evidence, especially color well M51. |
| Gestures | Visible Back/Save/details controls provide explicit actions. Dirty state disables swipe dismissal; confirmation protects navigation removal. Native gesture parity remains open. |
| Keyboard | App inputs and shared dismissal policy are reused. Invalid stable key requests focus. Ordinary single-line character loss remains an unresolved native blocker; source draft tests do not close it. |
| Accessibility | Inputs are labeled; errors move accessible focus where possible. M165 gives Save explicit name and pending label. Native form reading order, color naming and focus recovery remain open. |
| Motion | No custom form animation; disclosure rotates its chevron without a timed animation. Native transition/reduced-motion behavior remains unverified. |
| Content | Short form with Details disclosure rather than an always-visible technical key. Type description and inherited ownership are explained locally. |
| Search | No search within a short editor. Return to the collection for discovery. |
| Loading | Scoped reads show labeled settings progress. Record replacement resets through a workflow owner; stale loads cannot publish. Pending operations disable inputs/Save. |
| Recovery | First-load error has Retry; save errors retain draft and accessible summary. Permission changes retain read-only draft with Refresh access. Definitions refresh notice does not reset edited values. |
| Editing | Required name, key/color validity and dirty state gate Save. Workflow prevents repeated mutation and scopes completion to resource/focus; returned old success shows completion rather than reopening an editable draft. |
| Privacy | Context/read/mutation policy gates editability, with API denial handled separately. Inherited definitions require household management. Client gates do not replace server authorization. |
| Notifications | Expiration switch controls type tracking, not notification permission or delivery. Reminder settings are separate. |
| Media | No camera/library/audio operations. Color selection is metadata; no media permission is required. |
| Lifecycle | Focus/resource guards own discard and completion. Archive/restore/delete have scoped confirmation and pending locks. Native return, background and deep-link acceptance remains open. |

M165's named Save/pending test failed before migration; all54 customization mounted
tests, TypeScript and structural checks pass remotely. Existing inset assertion
now checks the containing column rather than custom button styling. M166 and M51
remain open; this is not a native editor acceptance pass.
