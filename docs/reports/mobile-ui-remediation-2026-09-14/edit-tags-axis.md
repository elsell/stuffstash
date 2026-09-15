# Edit tag selection: source review

Scope: S133, source de5d87b0 (PR146). This reviews all 24 axes; it does not
establish native acceptance. Tag selection is an in-place multi-selection task
within an asset edit draft. Color selection is a separate value choice. A new
navigation stack is not required for either task by the current product contract.

The decision standard is the platform-interaction review spec. The twelve-option
threshold comes from `specs/assets/asset-tags.spec.md`, not an Apple rule.
Apple's selection/input index was consulted; its JavaScript-only response did not
provide additional readable guidance for this pass. No unseen guidance is asserted.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Tags can be assigned, removed or created in the Edit form. M91 migrated completion actions to the native adapter; native geometry remains pending. |
| Navigation | Editing tags stays in the form. The route returns after successful Save or confirmed discard. Native return/focus restoration remains pending. |
| Selection | Existing tags expose selected/disabled state and a checkmark; pending tags expose selected state and can be removed. M41 preserves selected IDs atomically. Verify native announcement and identification of long similar names. |
| Modality | The containing sheet disables implicit dismissal and uses explicit dirty-draft confirmation. Native color picker presentation within the sheet still needs verification. |
| Layout | Metadata recovery and tags share the form scroll; the title and native footer remain outside it. Large-text reachability and keyboard/footer overlap are pending. |
| Adaptation | Tag choices wrap as a group, but names truncate to one line with a 180-point maximum. Name and 96-point hex field share a row. Narrow/iPad/large-text behavior is unverified. |
| Typography | Text scales, but selected and available names have a one-line cap. Do not equate truncation with overflow; test similarly prefixed long names and enlarged text. |
| Appearance | Tag tint uses a translucent background and theme text, rather than arbitrary tag color as text. Measure contrast, selected state and accessibility appearance modes on device. |
| Localization | Labels are English; normalization has Unicode-aware tests. Expanded labels, RTL ordering and actual locale sorting need review. |
| Imagery | Existing selected tags have a checkmark. Color swatches expose text names; tag identity is also textual. Pending-tag selection currently has no equivalent checkmark. Native distinguishability is pending. |
| Targets | Existing M40 review records 48-point minima. This does not prove visible, nonoverlapping hit regions after native layout. |
| Gestures | Explicit tag buttons and Save/Cancel avoid gesture-only completion. Scroll, keyboard dismissal and nested native color interaction remain pending. |
| Keyboard | Shared AppTextInput and scroll keyboard settings are used. The horizontal name/hex row and native footer need actual typing, traversal and dismissal checks. |
| Accessibility | Existing choices expose button roles and selected/disabled state. Overlong-name feedback has an alert role; invalid color feedback has a polite live-region declaration. Actual VoiceOver/TalkBack output and focus remain pending. |
| Motion | No tag-specific animation in EditTagPicker. Shared keyboard, sheet and system color picker transitions still need Reduce Motion verification. |
| Content | M94: every available tag is mapped into the form with no disclosure limit, contrary to the specified twelve-option initial list and selected summary. Large inventories push creation controls farther down. |
| Search | There is no dedicated existing-tag search. Typing a normalized existing name and adding it selects that existing tag, but is not a discovery/search interface. Assess realistic large inventories after disclosure repair. |
| Loading | Edit remains usable while tags load; missing data supplies an empty option array, without an explicit loading label. Existing asset selections remain in the draft. Delayed-load discoverability and interaction need review. |
| Recovery | M93 explains an overlong new name without clearing input. Metadata retry preserves the asset draft. Invalid color already has separate feedback. Native message placement/announcement remains pending. |
| Editing | Explicit Add stages a new tag; Save sends staged definitions and selected IDs. Partial creation is reconciled after Save failure. Unstaged name/color live locally in the picker and are not part of the route dirty check; draft-loss policy needs a separate decision and regression. |
| Privacy | Inventory-scoped query and asset-edit permission boundary are retained. This source pass does not replace adversarial boundary tests. |
| Notifications | No notification control in the picker. Notification interruption while editing remains part of lifecycle verification. |
| Media | No attachment acquisition here. Color selection is not a camera/library permission flow; no app-wide media acceptance is claimed. |
| Lifecycle | Route operation ownership guards async Save completion. Local unstaged input, backgrounding and inventory changes require native/draft lifecycle review. |

Sources inspected: AssetDetailSheets.tsx (EditAssetSheet/EditTagPicker),
AssetNativeActionSheetScreens.tsx (EditAssetForm), AssetTagDraftResolution.ts,
TagColorPicker.tsx, and the corresponding action-sheet/resolver behavior tests.
M93 has remote regression evidence; this pass adds no native verification.

Next: repair M94 with selected-tag retention and explicit native disclosure actions;
review unstaged tag input before choosing a draft-loss correction; capture long
names, large text, color selection, keyboard and Save/Cancel on native builds.

Follow-up candidate after PR146: M94 now has native disclosure commands with
selected extras retained when collapsed and natural ordering. The remote route
regression covers expansion, collapse and saved selection. This supersedes the
unbounded-list source observation above; native layout acceptance remains open.

Runner-only acceptance now covers fourteen choices with Tag 14 initially selected.
At the largest accessibility text size it reveals native disclosure controls,
selects Tag 13, collapses, verifies both selections and discards through the
production route. Captures distinguish expanded and collapsed states. Fixture
isolation checks, TypeScript and structural checks passed remotely; XCTest
execution and screenshot inspection remain pending. No Save or native visual
acceptance is inferred from preparation checks.

M95 follow-up supersedes the local unstaged-input observation: name and color now
belong to the route draft and participate in discard protection. Save cannot omit
them silently; the form explains Add/clear completion. Remote regression evidence
does not establish native input or message acceptance.

The native scenario also starts with an unchanged draft, types an exact new tag
name after keyboard readiness, dismisses the keyboard, verifies Save is disabled,
and uses Cancel / Keep editing. It checks retained input and the explanation's
visible bounds before Add tag clears the entry and enables Save. Completion still
discards the synthetic draft. Remote structural checks and critic review passed;
the Swift journey has not yet executed on a native runtime.
