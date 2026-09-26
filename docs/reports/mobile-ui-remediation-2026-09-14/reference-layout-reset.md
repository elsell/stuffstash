# Reference-led layout reset

The user rejects the composition of the bordered-button conversion. This confirms
an app-wide pattern-selection issue in the affected action families, not a need to
ask again whether each capsule looks wrong. New unrelated observations still need
confirmation. This document records the redesign and its verification boundary.

## What went wrong

The shared default changed the rendering of commands without changing their role,
placement or grouping. AssetDetailActions still makes a wrapping action strip;
SettingsActionRow wraps commands in another button inside a grouped row;
SettingsPickerRow delegates the entire row to a choice control whose Android
trigger is content-width. Native parts consequently retain an improvised whole.
There are 59 production TSX consumers/importers of NativeCommandButton; that is
an impact inventory, not 59 independent redesigns or tests.

Krug's scanning/hierarchy/conventions guidance makes the priority clear: users
should recognize the task and likely next action without deciphering equal-weight
controls. Apple's toolbar guidance assigns roles and groups by task/frequency;
its button guidance distinguishes prominence. Neither supports treating every
command as a separate bordered object.

## References and the limits of the evidence

- [Krug, Designing Pages for Scanning](https://www.oreilly.com/library/view/dont-make-me/0789723107/ch04.html):
  clear hierarchy, familiar conventions, distinct areas and reduced noise.
  This is the author's published work, not a consultation with him.
- [Apple Toolbars](https://developer.apple.com/design/human-interface-guidelines/toolbars):
  deliberate selection, familiar navigation/completion placement and logical
  grouping. Reviewed indexed official text; the direct page requires JavaScript.
- [Apple Buttons](https://developer.apple.com/design/human-interface-guidelines/buttons):
  visual prominence reflects priority; a native style alone is not a layout.
- [Apple Pop-up buttons](https://developer.apple.com/design/human-interface-guidelines/pop-up-buttons):
  flat exclusive choices, visible current value and predictable options.
- [Apple Files, iOS 26 workflow](https://support.apple.com/en-gb/102238):
  More exposes secondary commands; Move chooses a destination and completes with
  the system Done action. Documentation establishes the workflow, not pixel geometry.
- [Apple Contacts editing](https://support.apple.com/en-sg/guide/iphone/iph89a9c71d8/ios):
  read an object, enter Edit, finish with Done. Do not copy Contacts' communication
  shortcuts as arbitrary inventory commands; their frequency and meaning differ.
- [Apple Settings](https://developer.apple.com/design/human-interface-guidelines/settings):
  settings are coherent groups of choices, separate from an app's main work.
- User-provided ChatGPT screenshot: Back at leading edge, compact title/subtitle,
  grouped compose/More at trailing edge. This is direct visual evidence of the
  desired header. No claim was made to inspect the latest installed ChatGPT app.
- User-provided Facebook filter screenshot: aligned choice rows and a single
  prominent See items footer. Borrow its scanning hierarchy, not its custom skin.

## Copy the task structure

| Surface family | Reference to transfer | Stuff Stash composition |
| --- | --- | --- |
| Home/Browse headers | Supplied ChatGPT header + Apple toolbar grouping | Persistent compact chrome; body starts with content. Preserve Home Add/Notifications/Profile and Home-only inventory switching. Preserve Browse mode/search placement. No new row of duplicated body commands. |
| Details | Object inspection/Edit convention + Files secondary actions | Media and identity, then grouped facts and contents. Edit and More in the header. Move/Add photos/eligible Check out in More instead of a loose three-pill strip. Return stays contextual to an active checkout. Contents Add menu offers New item/Move items here. |
| Filters | Supplied Facebook hierarchy + Settings choice rows | One aligned label/current-value list. Flat values open a native menu; Tags opens selection. One prominent Show results action, quieter cancellation/reset. Preserve reachable keyboard actions. |
| Settings/account/sharing | Settings grouped rows + task forms | Whole-row destinations/commands; native value controls. Creation or submission has one primary action. Do not put a bordered button inside every command row. Invitation Copy belongs with its created link; Share uses system sharing. |
| Edit/New tag/New household/New inventory | Contacts Edit/Done + native form conventions | Named task, labeled fields, one completion owner, predictable cancel/return. Creation opens its own focused task; no inline form expansion. Retain unsaved draft and permission checks. |
| Move/New destination | Files move flow | Stable concise source context, destination list and search, one commit. New destination is a secondary toolbar command; its form has name, two-option kind and explicit create/cancel. Keep the existing sheet. |
| Inventory switcher | Grouped selection list | Compact household context, trailing Switch household, inventory rows with selection. Creation placed in its corresponding section/add menu, not a floating capsule among identity text. |
| Photos | Existing user-approved viewer direction | Preserve content-first full screen, top-right Close/More and continuous swipe paging. No bottom metadata/button tray; this reset does not reopen zoom or paging investigations. |
| Recovery/empty states | Local task feedback | Explain the problem beside the affected content; one Retry or relevant creation action. Do not spread repeated Retry/Back buttons across the screen. |

These are product adaptations of established patterns, not claims of exact Apple
screenshots or mandatory HIG layouts. Before native acceptance, compare rendered
screens at normal text to the actual reference composition; documentation or a
wireframe alone cannot prove the resulting layout is good.

## Implementation boundary and acceptance

Start with Details → Move and Filters/Settings rows, the clearest manifestations
of the confirmed problem. This requires moving/removing duplicate commands and
changing row composition, not changing NativeCommandButton's global default.
Inventory creation and photo lifecycle behavior remain intact. Audit the other
consumer families using the same role classification before changing them.

Source regressions must verify command reachability, pending/permission guards,
draft ownership and return behavior. Native review must cover populated and empty
Details, active checkout, contents Add menu, filter choice/apply/cancel, keyboard
entry and return, iPhone/iPad normal text and representative Android. Judge the
entire screen's hierarchy against the reference before accepting measured bounds.
Keep documented platform adapter limitations explicit. Do not add handmade visual
imitations of controls that an existing native adapter already supports.

Status: implementation candidate `32cc6465`. Source verification: 2,042 tests,
TypeScript and mobile structural checks pass. Code review findings (native width,
scalable iOS row text and circular styling import) were corrected. Android native
Details/More, Move and inventory selection layouts were inspected. Filter
availability selection/apply and the contents Add menu were exercised successfully;
menu row labels align with neighboring navigation rows. [Screenshots](evidence/reference-layouts/)
record these checks. iPhone/iPad
workflow and visual acceptance are pending; this is not a TestFlight release.
Existing release 140 checks do not establish acceptance of this new composition.

Design review clarified the retained Check out entry and scoped row/toolbar guidance
to these action families; it does not disallow tabs, breadcrumbs or contextual links.

Docs build reached site generation but failed because the pinned Pagefind Linux
binary could not be downloaded. This is an environment/dependency-fetch failure;
the source and mobile checks above passed.
