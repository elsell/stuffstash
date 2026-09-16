# Text-entry timing diagnostic

Prior native runs lose or reorder characters in both controlled and uncontrolled
ordinary fields. Disabling the accessory does not reliably remove the failure;
no-assistance cases passing do not isolate which assistance feature matters.

The upstream [controlled-input issue44157](https://github.com/facebook/react-native/issues/44157)
reports fast-entry cursor problems with correction/prediction and is marked fixed.
Its description says uncontrolled inputs were unaffected. Stuff Stash uses pinned
React Native0.83.6, and its uncontrolled baseline also fails. That report is a
research lead, not a diagnosis or justification for a dependency change.

The new diagnostic preserves all seven existing whole-string comparisons and adds
two explicitly paced cases: controlled and uncontrolled default-assisted fields,
with one character per XCTest call. Both native field and observed application
value must still equal the complete string. Screenshot names include `-paced`.
The fixed text-entry workflow includes the two additional cases.

This isolates sensitivity to automation cadence. It does not replicate physical
typing or establish whether UIKit, React Native, correction or automation is the
cause. A paced pass must not replace the original failure or become the production
acceptance method. Native compilation/execution remains pending; no production
input props, correction defaults, strings or baseline assertions are changed.
Remote structural checks pass (`/tmp/text-pacing-structural.log`) and code critic
found no confirmed issue. Neither result verifies Swift compilation or typing.

Follow-up: [iPad run350465](native-ipad-350465.md) compiled and executed both paced
cases successfully while the unchanged controlled whole-string cases still fail.
The seven original comparisons remain in the suite. This demonstrates cadence
sensitivity for that run, not a production fix or physical-typing acceptance.
The [phone result](native-phone-350465.md) now shows the same paced passes with
continued original controlled-input failures. Earlier uncontrolled failures still
prevent a controlled-only causal conclusion.

## Default native-field comparison

The next diagnostic adds a SwiftUI TextField with default text assistance, using
the same complete string and native/application-value assertions. The earlier
native system comparison was URL-specific and disabled correction/capitalization;
it could not isolate ordinary assisted entry. The new case does not alter any
production input, existing assertion or injection cadence. It is included in the
focused text-entry workflow and the full fixture suite. TypeScript and structural
checks pass on paul; critic found no blocker. Native compilation/execution remains
pending. Different implementation defaults and accessory attachment are potential
confounds, so even a pass will not establish a single-variable cause or physical
keyboard fidelity.

The350504 phone controlled baseline has a different retained final state:
[native and application values both show reordered text](phone-controlled-text-350504.png),
`Nativeraft name d`, rather than the requested `Native draft name`. Artifact
10429678833/testOrdinaryControlledTextEntry provides this screenshot. This remains
open; the later [onboarding address observation race](native-onboarding-350549.md)
does not establish that every text-entry failure merely needs a longer wait.
