# Locale, calendar and directional-layout audit

Source review September15. This document separates localized formatting from
translated product copy, calendar semantics and right-to-left layout. None proves
the others. Apple’s [picker guidance](https://developer.apple.com/design/human-interface-guidelines/pickers)
and [UIDatePicker documentation](https://developer.apple.com/documentation/uikit/uidatepicker/)
explain that native date controls use the user's locale/calendar/time zone;
the picker represents its selection with a calendar-agnostic NSDate. The domain
still stores a calendar date. A month-only period cannot be relabeled by
converting just its first day.

| Family | Inspected behavior | Result / remaining work |
| --- | --- | --- |
| Activity, checkout/return history, exact history details | Three formatters forced en-US. | M67 fixes date order and clock conventions through a shared formatter using the device locale/zone. Explicit British/US cases, actual en-GB runtime, invalid input and existing history behavior checks pass remotely. Native settings changes, 12/24-hour overrides and clipping remain pending. |
| Invitation expiry and sharing dates | Use default locale; sharing preserves invalid source value. | Native long strings, local zone and invalid invitation timestamp handling need acceptance. |
| History date-group headings and asset checkout summary | Use default locale. | Consistency with the changed timestamp consumers needs native review. |
| Expiration labels, month choices and group headings | Use default locale/calendar on synthetic Gregorian dates. | M68: alternate-calendar formatting can name the wrong month period. Keep exact-day instant formatting distinct from month-only interpretation. |
| Expiration status and refresh scheduling | Explicit en-US/en-CA formatter extracts or compares recipient-zone days. | These are internal comparisons, not US-only visible date labels. Retain authoritative recipient zone. Cross-midnight/DST and low-year edge cases require their own coverage. |
| Reminder time-zone chooser | Uses Intl for available/system zones and validation. | Searchability, localized names and long zone labels pending; do not confuse device zone with saved reminder zone. |
| Directional controls and custom Map geometry | Source contains physical offsets and custom chevrons; native back navigation is separately owned. | Run RTL with mixed-direction asset names, tags, URLs, breadcrumbs and Map panning. Physical offsets alone do not prove a failure, nor does automatic Yoga mirroring prove icon/gesture correctness. |
| Product copy and numeric inputs | Source copy is English; several numeric/year fields validate ASCII-shaped values. | Translation coverage and acceptance of localized digits must be explicitly evaluated. A formatting fix does not establish translated UI support. |

M68 evidence: using the existing formatter options on paul, stored Gregorian
`2028-02` formats as `Shevat 5788` for `en-US-u-ca-hebrew`, and the choice whose
stored value is January formats as `Tevet` from the synthetic January2020 date.
Those calendars' month boundaries are not interchangeable. Thai output uses a
Buddhist year while the numeric year input remains canonical Gregorian. This is
an input/period-semantics issue; confirm device calendar propagation separately.

The intended correction needs a coherent month-entry and month-display contract
that preserves the stored period and makes any calendar restriction clear.
Do not globally force US formatting or replace native exact-date controls as a
shortcut. M68 remains open pending that correction and its acceptance evidence.

M68 correction: shared month choices and month-only summaries now use localized
Gregorian names/years. If the locale selects another calendar, entry explains
the restriction and summaries append “Gregorian.” Exact-day presentation and
stored values are preserved. One baseline period-label assertion failed;14
formatter/field/workspace checks, TypeScript and structural checks pass on paul.
Critic found no implementation blocker; its date-versus-instant spec wording
correction is incorporated. Native calendar defaults, localized numerals and
clarification layout remain pending.
