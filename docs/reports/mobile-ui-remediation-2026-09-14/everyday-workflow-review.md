# Everyday mobile workflow review

Baseline: source4aec43d1, existing user observations and retained native evidence.
This is the first structural pass, not a completed runtime walkthrough of the
current build. Review at normal text before detailed-state or large-text work.

| Journey | Whole-experience judgment and next decision |
| --- | --- |
| Home → Browse | Keep Home's inventory choice and Add/notifications/profile hierarchy. Browse should preserve its own compact stable header and return context; verify the transition as a whole, not each icon separately. |
| Browse List ↔ Map | Confirmed structural inconsistency: List's SearchHeader puts BrowseSurfaceControl first in resultToolsRow, while Map places it after a flexing titleBlock in headerTopRow. List also owns it inside ListHeaderComponent; Map owns a separate header. The same native control therefore has different position and scroll ownership. Give the peer-view switcher one stable owner/anchor; do not merely match its width or recolor it. Verify repeated switching at top and after scrolling with query/filter state. |
| Browse → asset → back | Detail is a navigation destination. Preserve source view, query, filters and scroll position on return. Review title, photos, information grouping and action discovery together; per-command reachability alone does not establish good browsing. |
| Asset → Edit | Current iOS form combines name/description, expiration and tag management in a partial-height action sheet with large bottom actions. The iPad scroll failure is evidence of friction, but fixing geometry alone does not justify this container. Evaluate a dedicated full-height editor with coherent grouped fields and standard completion/cancellation semantics. Preserve drafts and return to the asset. Do not equate this recommendation with an Apple ban on editing sheets. |
| Asset → Move | Treat the task as choosing a destination, not filling out a second asset form. Start with current location, searchable hierarchy and clear destination selection; make creation an explicit secondary task. Judge whether a compact picker or full-height destination selector fits real inventories before retaining the current summary panels and sheet. |
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
