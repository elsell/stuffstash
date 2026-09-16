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
