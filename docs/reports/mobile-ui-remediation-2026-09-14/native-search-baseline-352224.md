# Unmodified search adapter baseline —35222427360

Source4420dc30ca90b769e77fa178a68a1bac7833b668, selection search-placement.
The native ordering transformation is not applied. Phone105205621758 passes6/7;
iPad105205622240 passes7/7. Phone Settings fails at line2211 asserting that the Search
button is hittable. Exact outcomes are in native-search-baseline-352224-results.csv. Retained
failure capture review is still needed to distinguish placement from missing UI.

Expiration keyboard/search/clearance passes on both with no redundant field tap
after Search activation. This supports the corrected single-activation acceptance
path; it does not prove the earlier AutoFill menu caused every keyboard failure.
All query/filtering and action-clearance requirements remained intact.

The unmodified baseline still fails a normal-text search consumer, while the
previous attachment-order experiment passed all phone cases. However, the runs
use different test revisions for Expiration and intermittent outcomes remain.
Run the ordering variant on this same revision before deciding to promote any
production patch; do not equate one green comparison with universal consistency.
No production dependency change has been made. The observer slept between checks
and collected the terminal result without rerunning failed jobs.
