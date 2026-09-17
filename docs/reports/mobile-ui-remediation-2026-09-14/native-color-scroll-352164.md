# Color scrolling comparison —35216493416

Source83dfabb96a7a4ac128e9a723c5364ca4e14d7156. Both devices pass8/9 focused
checks. Phone105186093622 fails the scrolling-disabled first-tap case while its
normal no-capture case passes. iPad105186093522 fails the normal no-capture case
while scrolling-disabled passes. Original opening, pre-tap capture, delivered
hit-region, RGB retention and disabled-state cases pass both. Exact outcomes are
in native-color-scroll-352164-results.csv; these are assertion results, not visual
acceptance of this run's screenshots.

Disabling scrolling is not sufficient to eliminate the miss: phone still fails.
It would also remove required editor behavior, so there is no basis for a production
scrollEnabled=false workaround. This comparison leaves other ScrollView behavior,
hosting and presentation mechanics intact; it cannot rule out the entire native
scroll host. Keep M51 open and retain contradictory samples instead of rerunning
until green. Further isolation needs native touch/presentation state or a distinct
native host, not repeated toggling of broad editor settings.

Ten fixture installer checks, mobile TypeScript and code critic review passed
before dispatch. No production code changed. A sleeping Bash observer collected
the terminal outcome; neither job was restarted. Released RGB/lock fixes remain
supported by this sample. The comprehensive normal-text-first audit continues.
