# Sharing cancellation: bounded native evidence

The pending invitation now offers a contextual Cancel invitation command opening
the existing native confirmation directly. This removes the single-command menu
involved in issue [#239](https://github.com/elsell/stuffstash/issues/239).
It does not establish that the entire recovery workflow passes on devices.

## Observations

- [First run](https://github.com/elsell/stuffstash/actions/runs/37127946651),
  source c9e4c84a6cd6052bce995df337451cb9f751ab5e: both iPhone and iPad
  reached the confirmation without the keyboard and invoked cancellation.
  Recovery stopped at an obsolete assertion expecting a raw fixture exception.
  The app displayed the current safe catalog error. Footer clearance passed.
  Retained [phone](phone-confirmation.png) and [iPad](ipad-confirmation.png)
  screenshots show the actual presentation.
- [Corrected run](https://github.com/elsell/stuffstash/actions/runs/37129618289),
  source 724ff9b7118a2d2ad2b04eff3502d7df5214e344: assertions now check safe
  error copy and reject raw exceptions for cancellation, copy and share.
  iPhone stopped on XCTest's confirmation-button hittability assertion.
  Its [screenshot](phone-corrected-run.png) shows the alert without keyboard
  obstruction; this does not prove tappability. iPad stopped before typing
  because XCTest could not obtain an interactive keyboard.

## Delivery decision and limits

The production tree was unchanged between these runs. Twenty-two Sharing
behavior tests cover confirmation, preserved draft, permission/scope ownership,
locks, cancellation and failure recovery. These support source behavior, not
native acceptance. The first native run supports the changed initial presentation.

Deliver the reviewed change after required checks; keep #239 open for complete
native confirmation and recovery acceptance. No third unchanged native run.
Retry, copy/share recovery and navigation return remain unverified on this
candidate. Device checks stay on the consolidated checklist, not a release gate.
This is synthetic fixture evidence, not a live invitation or connected API test.
