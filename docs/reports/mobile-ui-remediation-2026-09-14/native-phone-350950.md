# iPhone native results — run35095011021

Source403ef11c, merge093408914d8f3f229eb349eb91024bf08683ffdb, iPhone17.
Job104790103806 completed with **65/83 tests passing and18 failing**. This is an
older source revision than the current photo gesture/contrast candidate. iPad
fixtures remain active at this observation; iPad onboarding passes and phone
onboarding fails its exact-address observation (see native-onboarding-350592.md).

## Passing evidence

All three Add draft presentation journeys pass, including header-configured Add.
Add unfinished-tag retention also passes. Home header order/scroll and all three
nine-point action probes pass, as do Home tab return and Return recovery/cancel.
Both notice placement journeys now pass their live geometry and command checks.
Voice proposal lookup/search/return and close/reset retention pass. Native static
search comparison passes. These named results do not certify other consumers.

## Remaining normal-size observations

- Production Place and preconfigured Place fail waiting for the header Search
  button. The inspected production final capture shows an expanded **bottom**
  search field, retaining M207. Static route search passes in the same run; stable
  React options alone have not established consistent native placement.
- Sharing expects `audit@example.invalid` but receives `ait@example.invalid`.
  The final screenshot visibly retains the missing characters with the keyboard
  open. This consumer already uses an iOS default-value field; do not attribute
  every text loss to controlled React state or declare that migration sufficient.
- Controlled address and ordinary controlled name diagnostics retain missing
  characters, including the no-accessory name variant. Settings dirty-Back fails
  its interactive-keyboard precondition. These are separate observations until a
  shared cause is established.
- Move Here fails waiting for the candidate after retry. The later final capture
  shows Audit tent present, still unselected. This does not establish that it was
  present within the five-second assertion window, nor that selection worked.
- Footer and nested diagnostic sheets fail finding Diagnostic Tags. Inspected
  footer capture shows only title, blank body and bottom actions. Do not describe
  that fixture as passing merely because the footer remains reachable.
- Expiration accessibility audit reports clipped text. Its precise element and
  configuration still require artifact inspection.

Seven explicitly enlarged-text journeys also fail (asset recovery, command
height, detail commands, Edit metadata, Edit tags, footer appearance, Move Here).
Keep those tracked while addressing the remaining normal-size findings first.

## Retention and next steps

Log: `/tmp/native350950-phone.log` locally. Artifact10448184747 is retained as
`/tmp/native350950-phone.zip` on paul (about1.6GB). Only selected attachments were
extracted to `/tmp/phone350950-selected` on both hosts; no duplicate full xcresult
extraction. Inspect the corresponding iPad results when that live job completes;
continue the native search/text-entry diagnosis without relaxing acceptance.
