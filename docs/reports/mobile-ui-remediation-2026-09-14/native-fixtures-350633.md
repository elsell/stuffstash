# Native fixtures — run35063391045

Source65acfd8641151620efee8a1b09a19b09aae8fbe3, tested merge
bd089623f61c94f1e9f9c0e7f15d398717cc1673. Phone finishes57/79; the job fails.
iPad fixtures remain running at this checkpoint. Both onboarding jobs pass;
their separately inspected evidence is recorded in `native-onboarding-350592.md`.
These counts describe journeys, not product compliance.

## Normal-size evidence

All three independent Home action touch probes pass on phone, including Add's
capture that timed out in run350598. The new production tab-shell journey passes:
actions/accessory survive Browse-to-Home return, the bottom item remains reachable,
and no refresh indicator remains. Synthetic destinations and ports still limit
these claims; this is not live inventory or physical audio verification.

The standalone Home scroll check fails on nested duplicate `Recently changed`
StaticText nodes. Its log identifies both nodes under the same scroll heading.
Commit470e5b2a selects the first matching text container and retains displacement,
ordered/header geometry and stationary-header assertions. Remote structural checks
and critic review pass; native rerun is required.

Invitation acceptance passes the new long-name/access/accept-once/open-failure/
retry journey. Phone Place search passes this run, although the earlier run's
inspected bottom placement failed. Settings collection search still fails before
activation. Neither inconsistency closes M207. The paired preconfigured Place
comparison was added after this tested revision.

Account recovery fails its full-notice-below-navigation assertion. Its inspected
[final capture](evidence/phone-account-notice-350633.png) and hierarchy instead
show the entire notice at(16,126),370×70, below the header ending at116, inside
the402×874 window. The log retries both header/application lookup during the
predicate, then times out. This is contradictory observation evidence, not proof
of persistent overlap or a passed recovery journey. Preserve the assertion and
rerun before choosing a production correction. Sharing fails exact email entry before submission.
The Add tag journey reads `Te` immediately instead of `Tent`; this run predates
the exact-value wait justified by the previous final capture. Preserve the failed
result until rerun. Hidden-header Add readiness also still fails.

Color direct opening/clear/draft retention and single accessible name pass.
The separate color target check stops at a28-point accessibility frame; it does
not measure delivered touches. Phone voice location passes; the later M216
layout candidate and scoped-scroll correction are not exercised here.

## Remaining failures and evidence boundaries

Five text-entry comparisons fail: controlled address, controlled without
accessory, controlled without assistance, ordinary controlled, and ordinary
without accessory. These are not waived or relabeled by the Add observation fix.

Enlarged asset region, command height, detail commands, edit metadata, edit tags,
expiration overview, footer appearance and Move Here checks fail. Full and nested
sheet layout checks also fail. Normal-size findings remain the remediation
priority; this run does not supersede their open ledger entries.

Phone job104689596696; log `/tmp/native350633-phone.log`.
Artifact10435281653 retains screenshots, hierarchy and xcresult at
`paul:/tmp/native350633-phone.zip`. Account final-state capture is inspected;
other capture inspection and tablet results remain follow-up work.
