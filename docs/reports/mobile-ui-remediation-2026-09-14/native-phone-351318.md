# Phone native results — run35131892834

Source dfad17ab; tested merge21939a989d89e3f0e3a4bf10133f3dab63c8b073.
iPhone17 job104914725158 completed86 tests in3654.741 seconds:64 passes,
22 failures. This is not release acceptance. iPad fixtures remain active.

## New evidence

- Notification inbox read/navigation return passes. The separate delivered-touch
  test passes all nine center/edge/corner probes. This replaces the old invalid
  glyph-frame proxy with actual interaction evidence for the sampled phone state.
- Add photo preview/paging/removal and return passes with the corrected directional
  scroll. No production change is attributed to that test correction.
- Managed search enablement, header reconfiguration and adding a native header
  action pass. Production settings collection and preconfigured Place search
  still fail. A generic header update alone therefore does not reproduce the
  production failure in this sample.
- Sharing recovery, Home header, ordinary Place contents search and Voice proposal
  location search/retry/return pass.

## Remaining normal-size failures

The settings collection successfully invokes Add Tag, then lacks a hittable Search
button. Inspected capture `62C0617A-8C1E-47A1-9B82-728A87ACA6B5.png` shows a full
search field at the bottom instead of the requested header button. Hierarchy
`F1D38079-DB31-4CA1-A9FF-951C622E38E5.txt` places Search tags at(33,803),336×38.
Preconfigured Place has the same bottom-field geometry in
`70561620-384E-4B98-9D5C-BD56C17B3C1E.txt`. These remain native integration failures;
passing option-prop tests and the managed comparison do not clear them.

Ordinary color-picker opening fails again, while the separate target, delivered
region and single-accessible-name comparisons pass. M51 remains open.

Four text-entry comparisons lose or reorder characters: controlled address ends
`hple.invalid`; controlled name without accessory ends `Native dranameft `;
ordinary controlled name ends `Nadraft nametive `; seeded name without accessory
ends `N draft name`. The queued input-event trace run is needed to distinguish
native event delivery from committed React values; this run predates that trace.

Provider prompt Save recovery, seeded address without accessory and settings
editor Back recovery stop at the interactive-keyboard wait. Inspected provider
capture `155FE82E-054E-412C-A9E0-9D387CFCED16.png` later shows an empty focused field
and visible keyboard; its hierarchy identifies UIKeyboardLayoutStar Preview.
That later image cannot prove the keyboard was interactive during the wait, nor
clear the unexecuted save/recovery journey.

Add unfinished-tag disclosure reaches exact Camping input and disabled Save, then
fails an immediate assertion that its entry has disappeared after collapsing.
The final screenshot `B85E88A9-14D1-49DD-9881-25FDD32D0C32.png` still shows the
expanded field, but the subsequently collected hierarchy
`A15BBB68-BD97-41E7-8FDE-3C3FEB11C041.txt` contains collapsed guidance and no entry.
This mismatch is evidence of a later state transition, not proof of lost draft or
successful full recovery. The test must observe the completed transition before
continuing its reopen/preservation assertions. A test-only candidate now waits at
most five seconds for removal/guidance and exact Camping restoration. It retains
the Save, staging and clear checks and has not yet run natively.

Explicit enlarged-text failures remain deferred behind normal-size work. The
Expiration issue attachment8C0AAC7A-AAEE-4129-B284-D6347A5D273B explicitly reports
clipping at larger Dynamic Type sizes. FullSheet and NestedFullSheet comparison
failures remain separate from the passing shipped scroll-footer case.

## Retained evidence

Log: `/tmp/native351318-phone-complete.log` locally. Artifact10463884462 is
1,721,146,367 bytes, retained only on paul at `/tmp/native351318-phone.zip`.
Selected PNG/text attachments and manifest are under `/tmp/phone351318-selected`
on both hosts. No complete archive was downloaded locally. The iPad job was not
cancelled or restarted, and the later trace candidate remains queued.

Storage update (September 16): the downloaded full fixture ZIP copies for runs
35131892834 and35140471580 were removed from paul to recover temporary space.
GitHub confirms the original artifacts remain unexpired through September 30.
Retained selected captures, hierarchies and logs are unchanged.
