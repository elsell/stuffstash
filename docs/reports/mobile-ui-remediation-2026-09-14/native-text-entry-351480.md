# Focused text-entry evidence — September 16

Run [35148050909](https://github.com/elsell/stuffstash/actions/runs/35148050909)
at a01fc760 completed with **10/14 phone** and **11/14 iPad** comparisons passing.
These are diagnostic comparisons, not product acceptance. Ordinary CI35148054719
passed; the independent full native run35148054814 remains in progress.

The handled-tap correction works: all12 expected React Native traces per target
were exported, with zero dropped buffer entries. The two SwiftUI comparisons per
target intentionally do not use this trace. Exact-value assertions still fail;
trace capture did not convert failures into passes.

| Comparison | iPhone observed value | iPad observed value |
| --- | --- | --- |
| Controlled address | `hs://example.invalid` | `h//example.invalid` |
| Controlled name without accessory | `Ne draft name` | `Nve draft name` |
| Ordinary controlled name | `Native drmeaft na` | Full expected value; pass |
| Seeded name without accessory | `Ne draft name` | `Nraft name` |

Both targets pass the no-assistance controlled/seeded names, ordinary seeded name,
multiline name, paced names, seeded addresses with/without accessory, and SwiftUI
name/address comparisons. Passing diagnostic contrasts do not justify disabling
assistance, pacing all acceptance tests, or replacing every production field.

## What the traces establish

On phone, controlled address emits `h` at native event count1, commits `h`, then
emits `hs` at count2. No intermediate `t`, `tt`, or `ttp` text reaches this recorder.
iPad similarly emits `h` then `h/`. Counts remain contiguous.

The phone controlled name without accessory emits `N` at count1, commits `N`, then
emits `Ne` at count2. The seeded comparison has the same sequence. iPad emits `N`
then `Nv` for controlled, and `N` then `Nr` for seeded. Missing characters are
therefore already absent in the native change events received by JavaScript.

The ordinary controlled phone name instead progresses through `Native dr` at
count9, commits that value, then emits `Native drmaft na` at count10 and
`Native drmeaft na` at count11. This records misplaced text as well as omission.

These are JavaScript-received native events, not OS keystroke or UITextInput
instrumentation. Contiguous counts and zero buffer drops do not distinguish UIKit,
React Native event suppression, or XCTest synthesis. Commit adjacency establishes
ordering, not causation. Instrumentation itself can change timing.

## Source check and next boundary

React Native0.83.6 `RCTTextInputComponentView.mm` calls
`setDefaultInputAccessoryView` after props updates. A custom accessory ID reloads
input views when first responder; without an ID, unchanged accessory presence
returns early. `showSoftInputOnFocus` is updated only when its prop changes.
This warrants examining accessory updates but cannot explain the entire observed
failure class: comparisons without an accessory also fail. No dependency patch or
production behavior change follows from this observation alone.

Further localization should distinguish key delivery from text-change delivery,
while preserving ordinary typing and the unchanged product workflows. The full
native run remains necessary for Add, Return, settings and other real consumers.

## Retained evidence

GitHub artifacts10468633683 (phone) and10468513590 (iPad) contain result bundles.
The complete synthetic trace payloads and test/attachment mapping are preserved
in [phone JSON](evidence/phone-input-events-351480.json) and
[iPad JSON](evidence/ipad-input-events-351480.json), so their evidence survives
temporary-file cleanup and GitHub artifact expiry. Large archives stay on paul as `/tmp/native351480-focused-phone.zip` and
`/tmp/native351480-focused-ipad.zip`. Only small extracted traces are local in
`/tmp/phone351480-traces` and `/tmp/ipad351480-traces`; each has an index mapping
its test to its attachment. They are temporary copies; GitHub retention is finite.

Key phone attachments: address `6C979662-5094-4842-8E83-B8B1140640C9.txt`,
controlled no-accessory `72E22AFB-4420-4F56-BFB0-E6B3060374DF.txt`, ordinary controlled
`DAAEA565-4554-4193-8E91-6F93A390DB3E.txt`, seeded no-accessory
`B63E40BC-F02E-481B-A566-0240BC82B890.txt`.
Key iPad attachments: address `514C8A9E-6ED3-4FBE-A4D4-6ED8BB338BA8.txt`,
controlled no-accessory `80ABE38B-F30B-4AF8-B4EB-D35A0BD5A87F.txt`, seeded
no-accessory `C5E8031D-8E51-4173-90F0-E9F558DCD5C5.txt`.

Disk check after extraction:12GB free locally; paul4.9GB root and14GB `/tmp`.
No active build environment or untracked Android project was removed.

## Prepared follow-up

The runner-only recorder now accepts framework key-press events as well as text
changes and selections. Two recorder tests first fail on the missing key handler,
then pass alongside eight existing input/accessory tests on paul. They verify
ordering, no recording-driven render and bounded retention, not native delivery.
TypeScript, six fixture-preparation tests and the mobile structural check pass.
Code critic reports no confirmed blocker. Native execution of this additional
instrumentation is pending; the active full run is left untouched.
