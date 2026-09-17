# Native RGB rounding: candidate 5e192167

The installed @expo/ui55.0.17 source truncates normalized channels. macOS
CI35188426003/job105095410817 compiled its actual conversion expression and
failed: expected1, got0 for0.0039215686274509795. The one-line backport from
Expo commit dc32c26bd016c4187ada75d189c9d81f78a1060a adds nearest-integer rounding.

CI35188783384 succeeds. Its iOS dependency job105096587155 passes all9,472
installed-expression numeric checks, pod resolution and lock consistency.
The pnpm diff changes only the Expo patch identity; two ExpoUI pod paths follow
that identity. Existing Android patch additions/deletions are identical. Scoped
presentation tests, TypeScript and structural checks pass on paul. Critic found
no blocker. This accepts numeric conversion (M252), not complete UIKit editing.

Native35188788934 builds the candidate on both devices. Four focused cases run:

| Case | Phone105096535226 | iPad105096535411 |
| --- | --- | --- |
| Ordinary opening / close / preset / clear | Pass | App launch timed out before test body |
| Delivered touch region | Fails opening at center probe0 | All probes pass |
| Single accessible name | Pass | Pass |
| Target opens system picker | Pass | Pass |

The phone's center failure preserves M51; other passing cases do not erase it.
The iPad launch timeout does not demonstrate a color-control bug. These journeys
do not edit individual native RGB channels, so integrated channel retention
remains unverified despite the compiled numeric proof. No fresh captures were
visually reviewed for this run. Tests/logs are evidence only for their assertions.
The code is in draft PR157 and is not included in TestFlight114.1.

Shared consumers are Add tag creation, item Edit tag creation and Settings tag
customization through TagColorPicker/FullSpectrumTagColorPicker. The numeric fix
applies to their common iOS dependency; Android behavior is unchanged.

Sleeping shell observers waited for terminal jobs without restarting or polling
through model turns. Logs: /tmp/color-contract-red.log,
/tmp/color-contract-green.log and /tmp/color-rounding-native-JOBID.log.
Removed the inactive, regenerable /tmp/stuffstash-expiration-go-cache on paul,
recovering2.8GB; native evidence and source checkouts were retained.
