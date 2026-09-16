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

### Source follow-up: asset menu presentation

Inspection found a remaining source of native-header reconfiguration in the
production Place composition: `assetHeaderOverflowScreenOptions` was called on
every render, recreating its native menu callback even when presentation was
unchanged. The asset route now uses a presentation-memoized hook with current
committed handlers. Title, eligibility and disabled changes still update options;
loading/error clears the menu, and retained handlers cannot run disabled or
removed lifecycle actions or act after teardown. The existing search assertions
are unchanged. This removes unnecessary header updates, but a native rerun must
establish whether M207's placement changes.

A regression failed on callback-only option identity against the existing factory
wrapper, then passed with the hook. All46 focused hook/asset-route/iOS-menu tests,
TypeScript and mobile structural checks pass on paul. Logs:
`/tmp/asset-menu-{red,green,check,structural}.log`. This is source validation only.
Apple documents disabling toolbar integration to prevent bottom toolbar search;
our adapter already requests that setting. See
[search toolbar integration](https://developer.apple.com/documentation/uikit/uinavigationitem/searchbarplacementallowstoolbarintegration).

Combined validation after the menu change passes1,897 tests across296 files on
paul (`/tmp/asset-menu-full.log`), with TypeScript and structural checks also clean.

Review caught an unscoped retained-handler risk on same-title asset replacement.
A new regression failed before resource ownership was added; old handlers now
retire when `coreAsset.resourceKey` changes. The replacement handler's own removal
and teardown are checked independently. Critic confirmed the scope correction;
final combined source validation still passes1,897 tests/296 files, TypeScript
and structural checks (`/tmp/asset-menu-final-{full,check,structural}.log`).
