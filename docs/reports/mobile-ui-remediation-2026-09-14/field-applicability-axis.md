# Field applicability — all 24 source axes

S106 at1695cab3 plus M204. Reviewed CustomizationEditorFields, its parent editor,
existing draft/target recovery checks and the native choice adapter. This is field
applicability coverage, not certification of every field editor or native state.

| Axis | Conclusion and remaining acceptance |
| --- | --- |
| Task | Select where a field applies. A short exclusive native picker fits All assets/Selected asset types; eligible targets use in-place multiple selection. |
| Navigation | No new route for these choices. Parent editor owns Back, dirty-exit protection and return to collection. |
| Selection | Saved targets stay immutable; draft additions can be selected/deselected. M204 keeps unsaved expansion reversible without changing persisted-domain restrictions. |
| Modality | Native menu for applicability; no confirmation for draft selection. Parent Save is the persistence boundary. |
| Layout | Choices remain in the grouped scrolling editor. Shared content insets apply. Keyboard/scroll and long target lists need native acceptance. |
| Adaptation | Native picker label can wrap; targets are vertical rows. Tablet, narrow windows and enlarged text remain unverified. |
| Typography | Shared field labels/static values and wrapping native picker text. Native truncation/reading tests remain pending. |
| Appearance | Native choice adapter plus semantic settings rows; static values distinguish immutable fields. Light/dark runtime parity remains open. |
| Localization | English labels; target names are preserved. Inherited suffix and unavailable-count pluralization are explicit. RTL/long localized labels remain unverified. |
| Imagery | Check states supplement names. No photos or remote image dependencies. |
| Targets | Shared native choice and settings row hit regions; no new compact custom command. Actual native menu/row target acceptance remains required. |
| Gestures | Explicit selection does not require swipe/drag. Parent system Back/discard behavior remains independent. |
| Keyboard | Selection itself needs no keyboard. Parent name/enum editing can leave a keyboard active; runtime switching/focus still needs verification. |
| Accessibility | Choices expose current value, checked/disabled states. Saved targets use static text marked Existing. VoiceOver/TalkBack traversal and native menu speech remain pending. |
| Motion | No custom animation. Reduced-motion native presentation remains unverified. |
| Content | Eligible active types carry inherited context; unavailable saved targets appear only as a count, without hidden identifiers. Very large lists may need measured discovery improvements; no speculative route added. |
| Search | No separate applicability search in current editor. Scalability of target discovery remains a runtime/product question, not a proven small-list defect. |
| Loading | Parent loads definitions and eligible types; busy mutations disable draft choice. Incomplete/refresh state remains parent's responsibility. |
| Recovery | Missing saved targets stay retained; unavailable unsaved additions have aggregate removal. Read failure/retry must not erase the draft. Existing recovery tests cover these cases. |
| Editing | M204 separates persisted applicability from draft applicability. Switching away/back preserves targets and other edits; saved all-assets stays static. |
| Privacy | Eligible target query and parent permissions constrain editing. Read-only/inherited values do not expose mutation controls; backend authorization is unchanged. |
| Notifications | No notification permission or entry point. |
| Media | No camera/library/files/audio. |
| Lifecycle | Parent resource/focus workflow owns Save and discard. Retained native selection across blur/background remains an acceptance gap; no runtime claim from mounted tests. |

Two RED cases established missing in-place applicability choice. The mounted
roundtrip preserves a renamed field, immutable saved target and newly selected
target; a separate case rejects narrowing a saved all-assets field. All68 related
tests across two files, TypeScript and structural checks pass on paul. Code critic
found no confirmed blocker. Native menu
activation, keyboard return and disabled-open-menu timing remain unverified.
