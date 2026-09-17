# Phone fixture run350465 — terminal result

Job104638923406 completed with49/71 passing and22 failing. Sourceb6321dcb,
merge5a1e23d69395ff52f003e75d2fb61be3b7e6c472. Combined workflow is now terminal:
phone49/71, iPad58/71; onboarding phone1 pass/2 inapplicable skips, iPad3 passes.
This older native build does not verify later PR153 fixes.

Log `/tmp/native350465-phone.log`; artifact10428453726. The newer run35050407693
is already active at source802e4955; it is not a restart. Current71d54bb6 remains
queued behind it. Do not cancel a live run simply because newer fixes exist.

## Normal-size evidence

Four Add cases fail: navigation-stack name entry, rejected-save recovery,
configured-header name entry, and unfinished tag disclosure. Actual name values
include Nt name and Ne draft name instead of Native draft name; tag entry yields
Cing instead of Camping. These are not passing draft-retention scenarios.

Controlled text with and without the accessory also loses characters (Nve draft
name and Ne draft name). Both paced diagnostics pass unchanged native/app-value
assertions. Ordinary uncontrolled entry passes in this run; prior uncontrolled
failures remain relevant. No production typing fix is established.

Color open/clear fails waiting for the system Sliders control after activation.
The retained [capture](phone-color-activation-350465.png) and
[hierarchy](phone-color-activation-350465.txt) show the settings fixture still
visible, with no system picker. This is an activation failure, not merely a
different Sliders label. The accessible button is named twice (Choose any color)
and has a28×28 frame at346,360.7. No presentation-warning match was found in the
job log; absence of that warning does not identify the cause. The same open/clear
scenario passes on iPad in this run. Neither result establishes phone reliability.
The color target has a28-point accessibility frame. Home header has a36-point
frame. Frame size is not proof of the entire hit region; native hit diagnostics
and actual opening behavior remain separate.

Place search fails before typing, waiting for the Search button; settings
collection fails at the initial Search button's hittability. These are different
from the iPad Cancel-assumption failures corrected by46f23136. Do not apply that
iPad procedure correction as a claimed phone fix.

Sharing fails at an unhittable Cancel invitation menu action. Its later keyboard
dismissal/menu-target candidate is not in this source. Voice location fails at
the Back action after reopening the selected location; later explicit Back is
also not in this source.

FooterFullSheetLayout and NestedFullSheetLayout fail their existing comparison
fixtures. The earlier direct-scrolling production decision remains distinct from
these deliberately retained structural comparisons; do not label them newly
introduced production regressions without tracing their consumers.

## Enlarged-text backlog

Asset region recovery, command-height comparison, detail commands, Edit metadata,
Edit tags, expiration overview text clipping and Move-here recovery remain failed.
They are recorded without displacing normal-size remediation.

## Current source regression checkpoint

Source71d54bb6 passes all1,824 mobile tests across285 files, TypeScript and the
mobile structural check remotely on paul (`/tmp/mobile-audit-71d54bb6-full.log`).
Tracked mobile source/config/fixtures were checksum-compared and the only drift
(a renamed test description) synced before the full run. React act warnings
remain. This suite pass is not native acceptance or release completion.
