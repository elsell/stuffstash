# Phone native audit — run35112198520

Source1c2f8173, merge7b5627db6ebf613330b48bc20430ae3ad5d804c7;
iPhone17 job104851680499. Complete XCTest log:65/84 pass,19 failures,
3429.015 seconds. Job failure. Onboarding fails separately at help activation
(M240); its two iPad-only cases skip. Complete log:
`/tmp/native351121-phone-complete.log` locally. Artifact10457496721 is1,615,132,152
bytes, retained on paul as `/tmp/native351121-phone.zip`. Selected captures and
hierarchies are retained on both hosts in `/tmp/phone351121-selected`.

## Normal-size product work

- Add photo journey again fails its old root-menu return assertion after the photo
  operations; the newer retained-scroll acceptance correction is absent here.
- Add unfinished tag reads `Camp` immediately after typing `Camping`, but the final
  screenshot and hierarchy both contain exact `Camping`. This is evidence of an
  early observation, not a retained corrupted draft. The candidate test waits up
  to five seconds for exact equality before continuing the existing Save-blocking
  and disclosure-retention assertions. Native acceptance of that candidate is
  still pending. [Final capture](evidence/phone351121-add-tag-final.png).
- Ordinary native color opening fails, while target-opening, delivered-touch
  probes and single-name checks pass. M51 remains unresolved.
- Preconfigured Place search and Settings search again fail expected top search.
  Production Place and managed/static comparisons pass. This mixture does not
  establish that delayed registration alone causes the layout mismatch.
- Sharing reaches the older cancellation menu but Cancel invitation is not hittable;
  the new direct cancellation command is newer than this run.
- Normal Move Here rejection/retry and Add draft recovery pass.

## Diagnostics and deferred work

Controlled address reads `h//example.invalid`, controlled name without accessory
reads `N draft name`, and ordinary controlled name reads `Nativet name draf`.
These fail exact-value requirements; inspect final hierarchies before attributing
cause. Footer-full-sheet and nested-full-sheet comparisons fail but do not represent
the shipped route configurations. Expiration overview reports Text clipped; retain
its accessibility audit detail rather than assume which element or size failed.

Seven enlarged-text workflows fail: asset-region recovery, command height, Details
commands, Edit metadata, Edit tags, footer appearance and Move Here. Address normal
text findings first, as requested. None of these results validates newer inbox,
Sharing, Android-menu or test-observation changes.

The completed run remains a failed release gate. Newer run35121454700 at1a15ca11
has started normally and is monitored with a45-second sleep/change-only loop.
