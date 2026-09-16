# iPad native audit — run35104157358

Source14d7e06f, merge767d677340b3144968c155bb0f1a4b84d0548e1f;
iPad mini (A17 Pro), job104821415300. Complete XCTest log:77/84 pass,
7 failures,3331.436 seconds. Job failure. Onboarding3/3 pass separately.
Artifact10453515437 is retained as `/tmp/native351041-ipad.zip` on paul;
selected evidence is `/tmp/ipad351041-selected` on both hosts. Complete log:
`/tmp/native351041-ipad-complete.log` locally.

Normal-size failures: controlled address shows `h://example.invalid`, controlled
name without accessory shows `Native ft namedra`; Move Here candidate selector
still queries the old type at this revision; voice-location focused Clear collapses.
Three enlarged-text failures (Edit metadata, Edit tags, Move Here) remain deferred
behind normal-size work. The queued88fe7499 selector correction is newer than
this run and cannot be judged by it.

Ordinary color-picker opening, target probes, unassisted controlled-name comparison,
Sharing recovery, all Add presentations, photo paging/removal, Home actions/return,
Place search and the static focused-clear comparison pass. These do not erase
phone failures or establish complete cross-platform acceptance.

## Focused search clear comparison

The static route registers integratedButton search options with no app query
handlers. After focused Clear, its screenshot and hierarchy show Search in the
navigation bar, no search field and no keyboard. The test then reopens and types
Garage successfully. [Retained capture](evidence/ipad-static-search-focused-clear.png).

Apple's [integratedButton documentation](https://developer.apple.com/documentation/uikit/uinavigationitem/searchbarplacement-swift.enum/integratedbutton)
describes inactive search rendering as a button; it does not explicitly specify
this Clear transition. The transition is independent runtime evidence. Production
M232 can follow it: adjusted iPad acceptance requires restored locations, accessible
in-header Search, reopening, exact fresh query, selection and return. Phone remains
strict. Six fixture-preparation tests and structural checks pass; code critic found
no blocker after adding horizontal containment checks. The revised production
journey still needs a native rerun.
