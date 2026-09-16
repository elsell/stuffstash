# Phone native audit — run35140471580

Source a1b827e0, tested merge a14a86aaa33e5a8169feab902988bea15a9bac15;
iPhone17 job104943442022 completed69/86 tests,17 failures in3018.021 seconds.
Complete log: `/tmp/native351404-phone-complete.log`. Artifact10466544534
(1,579,029,776 bytes) is retained only on paul at `/tmp/native351404-phone.zip`.
Selected final captures/hierarchies are `/tmp/phone351404-selected` on both hosts.
The iPad fixture job was still active at this inspection; onboarding passes on
both devices as recorded in native-onboarding-351404.md.

## Normal-size evidence

- Controlled address retains `h//example.invalid`; ordinary controlled name retains
  `Nate draft nameiv`; controlled name without accessory retains `Nive draft name`.
  The ordinary-name final screenshot and hierarchy were inspected and agree.
- Preconfigured Add fails exact-name retention before save. Other Add entry
  comparisons, photo preview/removal and unfinished-tag retention pass. The latter
  includes the bounded collapse/reopen observation correction; it does not clear
  the distinct text-entry failures.
- M51 ordinary color opening still fails. M207 preconfigured Place search still
  renders at the bottom: inspected21DD9A55-B887-4508-8D75-B80C84E7B112.png confirms
  the expanded field rather than a top search icon. Settings collection search and
  the managed enable/title/action comparison pass this time; previous intermittent
  Settings failures remain evidence, not disproven by this pass.
- Notification read/navigation and delivered-touch-region checks pass. Home return
  recovery and provider prompt recovery pass. FullSheet and NestedFullSheet
  diagnostic layouts still fail; these are distinct from the shipped scroll-footer
  composition.
- Nine enlarged-text/accessibility cases fail. They remain recorded behind the
  user's normal-size priority; this run does not establish whole-app acceptance.

## Missing trace diagnosis and correction

No `input-events-*` attachment was exported. In the failed controlled comparisons,
XCTest tapped Capture input events and then repeatedly found no trace element.
Final hierarchies show the keyboard gone, the capture button present, and no trace.
The runner-only FixturePage uses ScrollView's default keyboard tap policy, which
consumes the first outside tap to dismiss the keyboard. This explains the failed
capture mechanism, not the text corruption.

The input-comparison container alone now requests handled taps while typing;
other fixture containers retain their default policy. After an accessible capture
command is tapped, absent output now fails explicitly instead of silently skipping
collection. The field and React-mirror acceptance observations remain saved before
capture; no assistance, value ownership or production field behavior changes.
Native acceptance of this diagnostic correction remains pending.

TypeScript, six fixture-preparation checks and the structural check pass on paul.
Code critic found no blocker in the tap/capture correction. The focused text-entry
selection also includes the four existing address comparisons so the next diagnostic
run can collect both address and name traces. A focused pass cannot clear the full
native release gate.

Storage update (September 16): the downloaded full fixture ZIP copies for runs
35131892834 and35140471580 were removed from paul to recover temporary space.
GitHub confirms the original artifacts remain unexpired through September 30.
Retained selected captures, hierarchies and logs are unchanged.
