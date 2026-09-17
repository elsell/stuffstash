# Native run35038625270 — terminal evidence

Source5c31d10172ceac6155a35f4d3875f57a95734747. Both onboarding jobs passed.
Phone fixture job104613563497 passed43/67; iPad fixture job104613563552 passed51/67.
These are terminal XCTest log results, not screenshot acceptance or current-head
verification. Logs: `/tmp/native350386-phone.log`, `/tmp/native350386-ipad.log`.

Normal-size failures still include input character corruption, Add readiness,
color-well/Home AX bounds, and place-content native search. iPad also fails settings
collection search; phone fails invitation cancellation hit testing and full/nested
diagnostic footer layouts. The Home Return error check now fails because its query
matches multiple elements, before it can establish visibility. That is a test
selector defect requiring correction, not evidence that the reveal fix passed.

The log's sparse hierarchy shows two identically labeled StaticText nodes in a
parent/child relationship. The candidate selects the first matching text container
and keeps the entire frame-below-header assertion unchanged. Swift compilation
and rerun remain pending; no successful visibility result is claimed.

AX frame dimensions alone do not establish touch bounds. Diagnostic footer cases
do not establish production filter failure. Enlarged-text failures remain tracked
but normal-size issues retain priority. This source predates the latest voice and
provider changes; it cannot validate them.
