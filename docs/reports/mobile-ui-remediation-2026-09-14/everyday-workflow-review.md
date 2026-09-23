# Everyday mobile workflow review

Baseline: source4aec43d1, existing user observations and retained native evidence.
This is the first structural pass, not a completed runtime walkthrough of the
current build. Review at normal text before detailed-state or large-text work.

| Journey | Whole-experience judgment and next decision |
| --- | --- |
| Home → Browse | Keep Home's inventory choice and Add/notifications/profile hierarchy. Browse should preserve its own compact stable header and return context; verify the transition as a whole, not each icon separately. |
| Browse List ↔ Map | Confirmed structural inconsistency: List's SearchHeader puts BrowseSurfaceControl first in resultToolsRow, while Map places it after a flexing titleBlock in headerTopRow. List also owns it inside ListHeaderComponent; Map owns a separate header. The same native control therefore has different position and scroll ownership. M260 candidate gives the peer-view switcher one persistent native-header owner; the content no longer owns duplicate switches. This is implemented with mounted ownership/state checks, but native fit and stability remain unverified. Verify repeated switching at top and after scrolling with query/filter state. |
| Browse → asset → back | Detail is a navigation destination. Preserve source view, query, filters and scroll position on return. Review title, photos, information grouping and action discovery together; per-command reachability alone does not establish good browsing. |
| Asset → Edit | Current iOS form combines name/description, expiration and tag management in a partial-height action sheet with large bottom actions. The iPad scroll failure is evidence of friction, but fixing geometry alone does not justify this container. M261 selects a full-height native editor with persistent native Cancel/Save, concise kind/type context and a single iOS keyboard inset owner. The candidate preserves draft and operation semantics; source tests pass, but native fit and the connected successful-save journey remain unverified. The new journey connects the real detail and Edit routes, real update command and cache invalidation to one isolated repository; it checks name persistence, Save with keyboard visible, updated detail, reopen and clean Cancel. Expiration/tag persistence and recovery retain separate acceptance requirements. Preserve drafts and return to the asset. Do not equate this recommendation with an Apple ban on editing sheets. |
| Asset → Move | Treat the task as choosing a destination, not filling out a second asset form. Start with current location, searchable hierarchy and clear destination selection; make creation an explicit secondary task. M262 selects a full-height destination picker for searchable inventory lists, persistent native Cancel/Move and concise current/selected context. Creation now requires an explicit action below existing choices. Source checks pass; native task fit and the connected return to updated asset details remain open. |
| Browse → Filters → results | Keep short exclusive choices in place. Searchable multiselect tags/location may justify a selection surface. Make active choices and return to results clear; avoid presenting every value change as another navigation task. Review the complete filtering sequence and its density before refining uncommon states. |

Next acceptance is a connected walkthrough, judging stability, task cost and
visual hierarchy before individual controls. Record actual phone/iPad build and
screens/recording; historical captures and source assertions do not prove current
runtime quality. Preserve the existing fixed text-loss release batch separately.

Apple supports segmented controls for peer selections and asks for deliberate
placement ([segmented controls](https://developer.apple.com/design/human-interface-guidelines/segmented-controls)).
Its [modality guidance](https://developer.apple.com/design/human-interface-guidelines/modality)
favors brief, focused modal tasks; [sheets](https://developer.apple.com/design/human-interface-guidelines/sheets)
can be appropriate for a bounded task. The concrete editor/destination choices
above are project recommendations to evaluate, not universal platform rules.
