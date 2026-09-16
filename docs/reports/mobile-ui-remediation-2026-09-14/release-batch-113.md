# Frozen release batch after TestFlight 112.1

Production cutoff: `33dfc002`. PR153 targets main; the prior shipped baseline is
TestFlight0.24.23(112.1). The next version/build identifier comes from the release
workflow, not this document's working batch name. No unrelated product work enters
this batch. Necessary batch corrections get an explicit updated cutoff.

Full native run35156794515 tests2043abb0, whose production mobile source, packages,
patches and app configuration match this cutoff. Later changes are diagnostics and
reports. Focused run35158457674 at6128c8f5 adds geometry observations, not a product
fix. Ongoing full audit coverage is not the release gate.

## Changed workflows and acceptance

| Group | Required acceptance |
| --- | --- |
| Home and inventory navigation | Add/Notifications/Profile order, scrolling, tab return, Return success/cancel/retry, inventory selection and load recovery |
| Add, Edit, Move and checkout | Complete entered draft retained, save/rejection recovery, dirty Back confirmation, current selection applied, no obsolete or departed action dispatch |
| Browse and expiration | Native search and value selection, last tag applied, full actions clear keyboard/accessory, date choice and return, no draft/query loss |
| Settings and customization | Search/Add, native choices, complete saved fields, dirty Back and archive confirmation, correct current color/reminder values |
| Sharing and notifications | Invitation cancellation and access-loss recovery, read-state ownership, navigation return, no action after access loss/departure |
| Photos and conversation | Paging/zoom/Close and removal recovery preserve draft; proposal edits and native commands retain current plan, location selection and recovery |
| Critical regressions | Onboarding/sign-in and complete address entry; authenticated/authorized boundaries; no crash, data loss, duplicate submission, stale command or blocked escape in these workflows |

Use named native journeys on phone/iPad and retained Android workflow evidence
for changed Android controls, together with source behavior/security tests and CI.
Source tests alone do not establish native layout acceptance. Existing native runs
may supply evidence for unchanged production code; do not rebuild merely because
an audit report changed.

## Current decisions

- M249 remains a batch blocker: required filter actions visibly overlap the phone
  keyboard accessory. Geometry diagnostics are pending; no offset guess is accepted.
- iPad last-tag application remains a batch blocker pending triage: geometry passes
  but the required selected-tag result is missing. See native-filters-351546.md.
- Calendar dismissal in351546 is an automation-target issue under investigation:
  the full-screen dismissal element's center lies inside the calendar. Correct the
  target geometrically and verify dismissal before claiming product failure/pass.
- M250 has iPad search-journey evidence and source regression tests; phone acceptance
  stops at M249 and is still required.
- Ordinary iOS color activation requires scoped triage because the batch changes
  its accessibility naming. Existing direct-well and delivered-touch passes are
  relevant evidence, but do not silently waive a contradictory required action.
- Raw controlled/seeded text comparisons, provider-free comparisons and alternate
  preconfigured navigation layouts are diagnostic evidence. They are not independent
  release gates. Complete entry/save in the actual changed Add/settings/onboarding
  workflows remains required; evidence of corruption there blocks the batch.
- Enlarged-text audit findings and physical assistive-technology coverage remain
  separately tracked. They do not require finishing the whole audit before this
  release. Any critical safety/data-loss finding is assessed independently of size.

## Ship condition and notes

Once this checklist passes, finalize PR153, preserve the explicit TestFlight notes
in its squash body, merge through the release workflow, and verify Apple VALID plus
notes readback. Continue the comprehensive audit after shipping; do not expand this
batch to absorb unrelated findings.

Draft tester notes, to reconcile with final accepted scope:

- More consistent native controls for inventory and item actions.
- Better draft protection and recovery when editing or reviewing changes.
- Improved photo preview and draft photo removal.
- Clearer invitation cancellation and access recovery.
- More reliable notification navigation and filter interactions.
