# Text-entry timing diagnostic — pending

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
