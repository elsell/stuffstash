# Frozen release batch after TestFlight 112.1

Production cutoff: `223d6d0a` (updated from `33dfc002` only for the required
iOS invitation email correction). PR153 targets main; the prior shipped baseline is
TestFlight0.24.23(112.1). The next version/build identifier comes from the release
workflow, not this document's working batch name. No unrelated product work enters
this batch. Necessary batch corrections get an explicit updated cutoff.

Full native run35156794515 tests2043abb0, whose production mobile source, packages,
patches and app configuration match the original33dfc002 cutoff. The subsequent
223d6d0a email adapter requires new native acceptance; other later changes are
diagnostics and reports. Focused run35159542174 at624bcf29 adds geometry observations, selected-tag
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
- iPad last-tag acceptance now passes in full351567 and focused351595, including
  the explicit checked-state observation in the latter. Retain the earlier351546
  failure as history; do not claim a production fix from the added observation.
- Calendar dismissal needs native acceptance of the corrected test target. The
 351595 outside tap hit the underlying Back command; c40e7965 restricts the target
  to non-command navigation space and verifies the Date range page remains open.
- M250 has iPad search-journey evidence and source regression tests; phone acceptance
  stops at M249 and is still required.
- Ordinary color activation passes the original native acceptance in full351567
  on phone and iPad. The focused correction selection retains this regression check.
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
run35162612814 was superseded while pending by release-corrections35165286470 at
966158e6. Full run35162604601 is now active but predates the email correction.
Its production source matches the original33dfc002 cutoff. This CI result
does not establish Swift test compilation or clear outstanding native gates.

## Current native decisions

Full351567 phone completes68/87, with50/55 required tests passing. Reviewed
settings capture shows bottom-search placement with a mismatched fixture title
(see the fidelity correction below); sharing
capture confirms reordered email. Expiration audit detail again concerns larger
text only. The iOS email correction223d6d0a has30 focused tests, TypeScript,
structural checks and critic review; its native Sharing journey remains pending.
The earlier native run remains relevant for unchanged workflows but does not
verify this new email adapter. See native-phone-351567.md.

Full351567 iPad completes80/87. All55 required journeys pass by log assertion,
including settings, sharing, color and filters. Four diagnostic input comparisons
and three enlarged-text follow-ups fail. Selected filter, settings and sharing captures are now reviewed; see
[native iPad evidence](native-ipad-351567.md). This does not certify every capture.

Settings acceptance correction:351567 used the internal audit-customization title,
not production Tags. The fixture title and native precondition are corrected. Keep
its native search journey pending; do not treat the mismatched header as sufficient
evidence of a production Tags defect. The existing integrated-button requirement
remains unchanged. See the fidelity section in native-phone-351567.md.

The release-corrections workflow selection reruns nine existing phone/iPad journeys
for filters, Sharing, Tags Add/Search, Add draft recovery and ordinary color use.
It uses ordinary fixture layouts and supplements this matrix; it cannot certify the
whole batch by itself. The selector test executes the actual workflow shell case,
checks exact membership and uniqueness, and verifies Swift methods exist. It fails
before the selector is added, then all nine preparation/selection checks and mobile
structural checks pass on paul. Code critic finds no blocker.

Current release-corrections dispatch:35165286470 at966158e6 is pending behind the
active focused run35159542174. Existing active native jobs are preserved. The full
current-head run35165285932 is also pending. Queued verification is not acceptance.

## Current correction source CI

CI35165285719 passes at966158e6744bc2c71687c5d93c7319dbbd2c94f4, including the
iOS email adapter, Tags fixture parity and focused selector. Required checks report
1,917 mobile tests across303 files,1,119 web tests across151 files and67 client
tests across8 files. Conversation browser journey, search PostgreSQL benchmark,
iOS dependency lock, web release image and self-host runtime also pass. Log:
`/tmp/ci351652-required.log`. Native acceptance is still pending; this does not
clear the phone filter overlap or verify the new iOS field on device.

The newer run35162604601 also passes all applicable phone/iPad onboarding cases,
including the alternate iPad form-column journey. Selected iPad captures are
reviewed in [onboarding regression evidence](native-onboarding-351626.md). The
intermittent help issue remains tracked; this pass does not establish a fix.
