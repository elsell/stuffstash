# Frozen release batch after TestFlight 112.1

Production cutoff: `223d6d0a` (updated from `33dfc002` only for the required
iOS invitation email correction). PR153 targets main; the prior shipped baseline is
TestFlight0.24.23(112.1). The next version/build identifier comes from the release
workflow, not this document's working batch name. No unrelated product work enters
this batch. Necessary batch corrections get an explicit updated cutoff.

Full native run35156794515 tests2043abb0, whose production mobile source, packages,
patches and app configuration match this cutoff. Later changes are diagnostics and
reports. Focused run35159542174 at624bcf29 adds geometry observations, selected-tag
checks and a corrected calendar dismissal target, not a product fix. Ongoing full audit coverage is not the release gate.

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

[Named native checks](release-batch-checks.csv) classify the87 fixture cases:
55 required workflow/regression checks,24 diagnostic comparisons and8 enlarged-text
follow-ups. The prior351480 results are included as historical evidence, not current
acceptance. The added Browse keyboard journey was not present in that older run;
later351546 results supersede old filter passes where they conflict. Separate
onboarding jobs, Android evidence, source/security tests and release CI still apply.
Normal-size control accessibility checks remain required; enlarged-text follow-up
does not excuse missing names or unusable controls at the default size.

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
- The prior expiration-overview automated audit reports only possible clipping
  at larger Dynamic Type sizes, with no element identified. Its
  [retained issue detail](evidence/expiration-accessibility-351480.txt) belongs to
  enlarged-text follow-up, not a demonstrated default-size release blocker. This
  mixed audit still supplies required normal hit-region, description and trait
  checks; triage individual reported issues instead of gating on its aggregate
  red status. Do not claim the enlarged-text issue is fixed.

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

## Onboarding checkpoint

Run35156794515 iPad passes the standard complete onboarding journey and landscape
adaptation. Its alternate form-column comparison stops at unchanged help expansion
(M240), before exercising its keyboard comparison. Keep that intermittent,
noncritical help issue in audit follow-up; it is not a newly demonstrated batch
regression. See [evidence and decision](native-onboarding-351567.md). Phone
onboarding also passes complete entry, help, keyboard dismissal and reachable
actions, with captures reviewed. Remaining required batch checks are still pending.


## Source verification checkpoint

CI35159545733 succeeds at624bcf292ae3637438d038ac294a8f61002de678, with production
code unchanged from the frozen cutoff. Required checks include1,915 mobile tests
across301 files,1,119 web tests across151 files and67 client tests across8 files.
The other CI jobs (iOS dependency lock, self-host runtime, web release image,
conversation browser journey and search benchmark) also succeed. Native workflow
acceptance remains separate; this does not clear the filter failures. Full required
check log: `/tmp/ci351595-required.log`.

Current correction checkpoint: CI35162604544 succeeds atc40e7965, including the
calendar dismissal test correction and retained audit reports. Focused native
run35162612814 is queued behind35159542174; full run35162604601 is queued behind
35156794515. Production remains identical to the frozen cutoff. This CI result
does not establish Swift test compilation or clear outstanding native gates.

## Current native decisions

Full351567 phone completes68/87, with50/55 required tests passing. Reviewed
settings capture confirms the M207 bottom-search placement failure; sharing
capture confirms reordered email. Expiration audit detail again concerns larger
text only. The iOS email correction223d6d0a has30 focused tests, TypeScript,
structural checks and critic review; its native Sharing journey remains pending.
The earlier native run remains relevant for unchanged workflows but does not
verify this new email adapter. See native-phone-351567.md.

Full351567 iPad completes80/87. All55 required journeys pass by log assertion,
including settings, sharing, color and filters. Four diagnostic input comparisons
and three enlarged-text follow-ups fail. iPad capture review remains outstanding.
