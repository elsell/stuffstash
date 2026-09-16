# Custom field type and option entry

S105/S107, source 9c2c7246 plus M188. Reviewed CustomizationEditorFields,
CustomizationEditorScreen, native choice adapters, editor validation/commands and
existing mounted journeys. This covers the nested controls, not every editor state.

The task is choosing one of six field types, then editing a list of enum values.
A native menu fits the short exclusive type choice; adding/removing draft values
belongs in the existing form. Saved type and enum values are immutable by contract.
This is project task/pattern judgment. Apple's [picker](https://developer.apple.com/design/human-interface-guidelines/pickers)
and [text-field](https://developer.apple.com/design/human-interface-guidelines/text-fields)
pages returned JavaScript shells during this review; no unobserved wording is
asserted as a requirement.

| Axis | Source result and remaining acceptance |
| --- | --- |
| Task | In-place exclusive type choice; enum entry edits the form draft. |
| Navigation | Neither control pushes a new screen. Parent owns back/discard. |
| Selection | Native menu includes all six types and current value; allowed-value and disabled guards remain. Saved type is readable text. |
| Modality | Only the native choice menu opens; option editing remains inline. |
| Layout | Vertical rows and full-width option field; native Add/Remove commands. Actual keyboard/command fit remains pending. |
| Adaptation | No fixed screen height in controls; iOS choice stacks at accessibility sizes. Normal-size phone/iPad captures still required. |
| Typography | Semantic text roles are declared only in part; English labels wrap through RN Text. Large-text review remains later. |
| Appearance | Semantic palette and native choices/commands; disabled contrast requires native verification. |
| Localization | English labels; existing key normalization uses ASCII keys, not arbitrary localized option labels. M188 prevents an unusable normalized draft disappearing. RTL remains unverified. |
| Imagery | N/A: these controls have no authored image or icon content. |
| Targets | Native adapters own command/choice targets; measured bounds and menu hit tests pending. |
| Gestures | No gesture-only task. Parent scrolling and dirty-exit behavior still require native acceptance. |
| Keyboard | Option field uses shared input; parent adjusts keyboard insets and supports dismissal. Ongoing native text-entry failures prevent a typing acceptance claim. |
| Accessibility | Picker names current value; option field has a label and M188 error hint/live text. VoiceOver/TalkBack reading and announcement not verified. |
| Motion | No custom motion; native menus/keyboard remain system behavior to inspect with reduced motion. |
| Content | Existing enum values are labeled Existing; unsaved values expose Remove. Empty enum explains the missing option. |
| Search | N/A: six type values and locally authored options do not require a search route. |
| Loading | Parent owns initial loading; pending save/lifecycle disables changes. |
| Recovery | M188 preserves rejected empty-normalized or duplicate input and explains correction. Successful Add clears it. |
| Editing | Pending option participates in dirty protection and enum Save validation; switching type preserves dormant options and non-enum create omits them. |
| Privacy | Parent gates mutability by context/scope/inherited ownership. Controls do not perform network requests; this is not backend authorization evidence. |
| Notifications | N/A: no system notification entry or delivery in these controls. Inline validation is feedback, not notification setup. |
| Media | N/A: no acquisition or playback. |
| Lifecycle | Parent owns resource/focus completion and guarded exit. Retained native events across disappearance still require runtime coverage. |

M188 reproduced loss of both invalid and duplicate draft text before implementation.
All 70 related mounted/model checks plus TypeScript and structural validation pass
on paul (`/tmp/enum-entry-green.log`); the screen suite emits React act warnings,
so its pass is not a warning-free run. Code review found no blocker.
Native validation placement, keyboard focus, typing fidelity and assistive feedback
remain open; source review does not close those findings.
