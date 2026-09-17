# Phone native audit — run35156794515

Job105001197879 at source2043abb0 (tested merge2209ff1c347187e92eef215f965b5fe68e040df2)
completes68/87 tests in3570 seconds. Production matches the frozen release cutoff.
The named release mapping yields50/55 required passes,17/24 diagnostic passes,
and1/8 enlarged-text follow-up passes. These are log/assertion results, not a
claim that all captures have been reviewed.

## Required failures to triage

- Browse and expiration keyboard journeys fail complete footer/accessory clearance,
  reproducing M249. Full runs do not include the focused geometry probe.
- Sharing complete email entry receives `aexample.invalidudit@` instead of
  `audit@example.invalid`. This is an actual workflow failure and requires triage;
  it cannot be waived as one of the isolated input comparisons.
- Settings collection Add succeeds, then Search is not hittable at source line2065.
  This fails before search text entry. Review capture/hierarchy before attributing
  the failure to header state or a missing control.
- Expiration overview reports Text clipped. Inspect the issue attachment to decide
  whether this is again enlarged-text-only; aggregate failure alone is insufficient.

Add draft recovery and its navigation comparisons pass. The ordinary color picker
opens within its original5-second criterion and passes clear/draft retention; this
run does not need the late-presentation fallback. Historical activation failures
remain tracked rather than being represented as universally resolved.

Artifact10474460392 (1.68GB) is being retained on paul as
`/tmp/native351567-full-phone.zip`. Complete local log:
`/tmp/native351567-full-phone.log`. Selected artifact review is recorded below.

## Selected artifact review

The [settings capture](evidence/phone-settings-search-351567.png) and
[hierarchy](evidence/phone-settings-search-351567.txt) show Add in the header but
Search tags as a persistent bottom field at33,803,336,38. Search is not absent; its
placement differs from the requested integrated header button. This extends M207
to this required settings workflow; do not change the test to accept the bottom
field without resolving the intended product interaction.

The [sharing capture](evidence/phone-sharing-entry-351567.png) visibly confirms the
reordered email, with the insertion cursor in its middle. The
[hierarchy](evidence/phone-sharing-entry-351567.txt) agrees with the assertion.
No typing/provider root cause is established by this screenshot.

The [expiration issue](evidence/expiration-accessibility-351567.txt) again reports
possible clipping only at larger Dynamic Type sizes, without identifying an
element. Keep that issue in enlarged-text follow-up, while normal trait/name/target
acceptance remains required. This is not evidence of a new default-size blocker.

Sharing candidate: the iOS email field now uses the existing seeded SwiftUI
TextField pattern, with email keyboard/content type and editing disabled while
creation is pending. Existing scope/reset revision and command ownership remain.
The new mounted seed/lock/reset test first fails for the absent adapter;30 scoped
sharing/guard/adapter tests, TypeScript and structural checks then pass on paul.
Critic reports no confirmed blocker. These checks do not prove native typing or
keyboard dismissal; the unchanged native Sharing journey remains required.

## Settings fixture fidelity correction

Further inspection found the capture title is the unregistered fixture route
`audit-customization`, while production registers this collection as `Tags` in
`src/app/_layout.tsx`. Header width therefore differs materially. The bottom field
is observed, but this run alone does not prove that production Tags has the same
placement failure. FixtureLayout now registers Tags, and the native journey asserts
that title before Add/Search. The integrated-button and result checks are unchanged.
A structural parity test fails before correction and passes afterward; remote
TypeScript, eight fixture-installer checks and structural checks also pass. Critic
finds no blocker. Native rerun remains required. This does not explain the separate
Place comparison, whose fixture already uses the short Place title.

## Additional changed-workflow visual review

Four retained phone captures add visual evidence to the passing native journeys:

- [Home after tab return](evidence/phone-home-return-351567.png): Add,
  Notifications and Profile remain visible in that order over scrolled content.
  The voice accessory and tab controls remain distinct; no refresh spinner appears.
  The long inventory name truncates within its own control.
- [Return details after rejected save](evidence/phone-return-error-351567.png):
  the entered “Returned clean” draft and inline error remain visible inside the
  sheet, with Cancel return and Save both unobscured.
- [Surviving draft photo](evidence/phone-photo-preview-351567.png): the remaining
  image, 1 of 1 count, Close and Remove controls remain visible after removal.
  This checks this state only; the dark status glyphs against the dark viewer
  are tracked as M251 outside this batch, not evidence of blocked photo commands.
- [Native color picker](evidence/phone-color-picker-351567.png): the system color
  panel and Close control are visible after ordinary activation.

These are light-appearance simulator captures from the pre-email-correction
build. They support unchanged workflows; they do not verify the new email adapter,
all appearance settings, physical assistive technology, or the entire app.

## Draft and proposal visual review

The [rejected Add draft](evidence/phone-add-rejected-351567.png) retains the complete
“Native draft name” beside the save error, with Close and Save visible in the header.
The [returned conversation proposal](evidence/phone-voice-return-351567.png) retains
its item and inventory-root selection, with Approve/Cancel unobscured. After the
location journey, the [same proposal](evidence/phone-voice-location-351567.png)
shows Garage / Garage bin with those commands still visible. These captures support
the passing normal-text draft/recovery tests; they do not establish physical voice
input or every interrupted-request state.

The Move successful-retry capture was also inspected, but it shows only the fixture
launcher after return. It supports the terminal destination, not the appearance of
the intermediate Move selection/error state. That journey's assertions remain the
evidence for selection retention and retry; do not describe this capture as a
visual review of the Move sheet.
