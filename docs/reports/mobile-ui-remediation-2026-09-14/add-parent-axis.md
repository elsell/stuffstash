# Add placement search — 24-axis source review

S086 at716f23b1 plus M200. Reviewed AddAssetScreen's ParentPicker/ParentOption,
AddAssetResolution, AddAssetInitialParent, useParentCandidates and ParentLookupQuery.
The search is an inline disclosure within the unfinished Add draft, with selection
and quick place creation. It does not require a second navigation stack. Options
include items (with explicit promotion to container), places, containers and root.

| Axes | Evidence and outstanding acceptance |
| --- | --- |
| Task, navigation, modality | Placement is part of the Add draft. Opening/closing the disclosure does not save the item or navigate. Creating a place is an independent mutation; abandoning the item does not delete it. Native disclosure/keyboard fit remains open. |
| Selection, editing | Selecting a candidate closes the disclosure; typing clears the explicit selection. Blank is top level; a known exact title can resolve during Save, while unresolved text blocks Save with correction guidance. M200 stops the unresolved header from falsely claiming top-level placement. Draft restoration retains the last selected identity/context. |
| Layout, adaptation, typography | Parent disclosure minimum56 points, rows minimum48. Nested results are bounded at260 points inside a340-point panel; labels wrap. This is not proof the nested scroll or keyboard layout fits a narrow phone/iPad. Existing native text-entry failures remain open. |
| Appearance, imagery | Semantic palette, selected border/check, optional root inventory icon. Location metadata is text, not dependent on images or color. Native-looking density/materials and disabled contrast need device review. |
| Localization | English placement/promotional copy; case-insensitive trimmed matching uses locale lowercasing. Hierarchical subtitles use slash-separated text. RTL and long names are not certified. |
| Targets, gestures, accessibility | Disclosure has expanded/disabled state; candidates have selected/disabled state; labels and metadata are descendant text. Search has an explicit name. Verify VoiceOver grouping/order, checkmark announcement and real target bounds; source minima alone are insufficient. Nested scrolling has explicit selection alternatives. |
| Keyboard, motion | Search autofocus schedules scroll-to-end in the outer form; both scroll containers support keyboard dismissal and handled taps. Verify this reveals the focused control instead of overscrolling when More details is expanded. Reduce Motion behavior of that animated scroll remains unverified. |
| Content, search | Query is debounced250ms, scoped by inventory/service, canceled when superseded/hidden. Exact matches are prioritized; visible results are capped at6 (5 when blank). This is a suggestion list, not a complete inventory browser. Root remains available. |
| Loading, recovery | Quick creation waits for known results; unavailable suggestions expose Retry. Selection retains last-parent context independently of transient search results. New-place creation failures preserve text and expose shared inline error recovery. |
| Privacy | Shared server query and draft ownership carry inventory/principal context. Candidate lookup alone is not authorization to create/move; those commands enforce the boundary. This review does not certify backend isolation. |
| Notifications, media | No notification or media operation is owned by placement selection. Add's separate photo work locks the form to avoid competing mutations. |
| Lifecycle | Busy Add prevents route removal; idle draft is retained in the scoped in-memory store. New queries cancel obsolete suggestions. Background/process death, navigation interruption and physical keyboard behavior require native evidence. |

No native pass is inferred from this review. M200 is a status-copy correction,
not a redesign of placement search or a change to exact-match Save semantics.

All28 related tests across4 files, TypeScript and structural checks pass remotely
on paul (`/tmp/add-parent-label-green.log`). Code critic found no blocker.
