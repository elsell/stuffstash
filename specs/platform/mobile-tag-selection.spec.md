# Mobile Asset Tag Selection

Status: M268 source integration complete; native acceptance pending.
This is outside the frozen M260–M264 release and M265–M267 candidates.

## Task and platform pattern

People assign several existing inventory tags while creating or editing an asset.
The inventory can contain many tags with long names. Add, Edit and Filters must
share recognizable search, selection feedback and ordering. Existing Add/Edit
chip grids are a known implementation gap, not the desired reference pattern.

Apple recommends a list or table for very large choice sets and predictable
ordering: [Pickers](https://developer.apple.com/design/human-interface-guidelines/pickers).
Selection feedback must distinguish toggling from navigation:
[Lists and tables](https://developer.apple.com/design/human-interface-guidelines/lists-and-tables).
These principles support a searchable multi-selection list. The concrete choice
of a selection view is a project decision, not a universal Apple requirement.
Small single-value choices elsewhere must retain their in-place native menus.

## Form and selection behavior

- Add and Edit show one Tags field with a concise selected-tag summary and count.
  Unselected options must not occupy the asset form. Long names must remain
  available in the selection view rather than being irretrievably truncated.
- Opening Tags presents a native navigation-owned selection view with search,
  alphabetically ordered checkmarked rows and a stable selected-count summary.
  Use the existing filter selection-row semantics and native search adapter;
  inspect their consumers before changing shared behavior. Do not add nested
  scrolling option grids or a custom dropdown imitation.
- Selection changes stay in a local selection draft until Done returns them to
  the owning Add/Edit draft. Cancel, back or interactive dismissal discards only
  this selection visit. Saving the asset remains the outer form's responsibility.
- Search matches tag names, preserves all selected IDs when results change, and
  distinguishes no matches from no tags. Provide a way to review selected tags
  without clearing the user's query or silently dropping hidden selections.
- Assigned IDs absent from the current tag list remain selected until explicitly
  removed. Selected review exposes these as unavailable tags with removal controls;
  it must not claim no selection or expose raw IDs as names.
- Keep optional New tag creation in the owning form as specified by M264/M267.
  Pending new tags and unfinished creation input survive an existing-tag
  selection visit. Selection must not create or mutate inventory tags remotely.
- Permission loss, pending asset writes and inventory/tenant changes invalidate
  selection callbacks through current committed scope. Never apply an old
  selection to a new asset draft or a different inventory.
- Use one shared selection model/presentation for Add and Edit; their ports,
  persistence and draft lifecycles remain independently owned. No business logic
  belongs in native navigation or transport adapters.
- The root presentation provider carries only the current selection visit. The
  Add/Edit field owns that visit and its committed callbacks. Removing the field
  or changing its scope invalidates the visit. Native back/dismissal cancels it;
  Done applies once and dismisses. Provider updates must not reset local search
  or selection, and must not repush a route already open.

## Acceptance

Review Add → choose two tags → search elsewhere → review selected → Done → save,
and Edit → change selection → Cancel → reopen → Done → save/reopen detail.
Include selection past the first screen, long labels, no matches, no tags, failed
asset save, restored Add draft, pending new tags and inventory-scope teardown.
Use meaningful mounted tests first, then native phone/iPad and Android workflows.
The existing Filters tag journey is a representative shared-row/search regression;
do not make every surface/axis cell an individual test. A component-native control
or green unit test does not prove the end-to-end task fits platform conventions.

Native fixtures must include existing choices beyond the initial viewport in both
Add and Edit. Add acceptance includes named asset draft retention through Cancel,
selected-review search, Done, and a rejected save; this must not rely solely on
Edit sharing the selection component.
