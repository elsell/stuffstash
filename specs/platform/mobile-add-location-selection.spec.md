# Add Item Location Selection

Status: M269 native Add destination integration implemented; native acceptance pending.
Keep separate from frozen M260–M264 and tag-selection native acceptance.

## Problem and pattern

Add currently expands a custom searched, nested scrolling list inside its form.
Typing replaces parentQuery and clears parentAssetId before a choice is made;
collapsing the list does not restore the prior selected destination. Native
Android review confirms the inline interaction and nested presentation. It did
not reproduce a hidden search field; do not report that as a runtime defect.

Choosing a containment destination can require searching many descriptive,
hierarchical options. Use a native navigation-owned single-selection list with
current-location feedback, consistent with the destination-first Move task.
This project choice follows Apple's guidance on
[selection](https://developer.apple.com/design/human-interface-guidelines/pickers),
[search](https://developer.apple.com/design/human-interface-guidelines/searching)
and [lists](https://developer.apple.com/design/human-interface-guidelines/lists-and-tables).
It is not a general requirement to navigate for small flat choices.

## Behavior

- Add retains a compact Put in disclosure row showing its chosen value, using
  the existing SelectionRow; supporting path text stays below it.
  Opening it never changes the draft or creates an asset.
- The selection visit owns its query. Search updates results without clearing or
  changing the parent held by Add. Preserve the selected destination when hidden
  by the query. Show inventory top level explicitly and expose path/type context.
- Tapping an eligible destination, including top level, applies it and returns
  to Add. A single choice needs no second Done confirmation. Back, Cancel or
  interactive dismissal leaves the original destination and the rest of Add intact.
- Use the existing native search adapter and one scroll owner. Remove the nested
  result scroller and whole-form scroll-to-end callback for parent-search focus.
- Loading, no matches, failure and retry are distinct. Disabled destinations
  retain their explanation. Do not offer creation until the current lookup has
  resolved and confirmed no equivalent existing destination.
- Creation is an explicit secondary New place task, not an implicit consequence
  of typing or saving Add. Reuse the existing creation command and operation
  lock. Creation failure stays visible with its draft in the active task; retry
  does not duplicate a completed creation. On success select the created place
  and return to Add. Explain the independent creation boundary: canceling Add
  later does not delete a place that has already been created.
- Keep the original asset draft, photos, staged tags and unfinished tag entry.
  Restored legacy parent drafts remain supported; new search input is not saved
  as a destination until the user chooses one.
- Current committed scope/permissions own all callbacks. Prevent new visits from
  blurred owners; opening a visit must not invalidate it just because Add blurs.
  Inventory/principal changes and unmount reject late selection/creation results.
  Preserve pending-operation dismissal protection and authorized command ports.

## Acceptance

Mounted regressions first: existing destination → search another → Cancel;
choose another → save; top-level choice; failed lookup and retry; explicit create
failure/retry; owner teardown and retained callbacks; other draft fields retained.
Then normal-text native Add → choose destination → back/choose → rejected save
and retry, including keyboard and a result beyond the first viewport on phone,
iPad and Android. Inspect the whole form-to-selection transition, not just row
props. Shared Move lookup/creation behavior gets representative regression checks.

The integrated workflow passes 80 focused Add and shared asset-action tests,
TypeScript and mobile structural checks on the remote validation host, plus ten
fixture-preparation tests. Coverage includes cancel/existing/top-level/create
choices with rejected item save, permission loss, blurred opening, actual route
removal and pending-operation protection. Code review found no source blocker.
Native Android acceptance and iPhone/iPad acceptance remain separate requirements;
these checks do not prove device navigation, keyboard or visual behavior.

Android connected acceptance on source552e7c1b reached search/cancel, direct
selection, rejected item save and creation failure/retry, but crashed on creation
success return with ScreenStackFragment added into a non-stack container during
header update. Keep route removal protection installed throughout the visit and
its exit, following HomeReturnDetailsRoute; read current task state in the removal
callback. A live task cancels only when unlocked; a completed task dispatches the
original removal. Do not toggle native removal configuration in the completion
render. Reverify the connected creation-success return before native acceptance.
