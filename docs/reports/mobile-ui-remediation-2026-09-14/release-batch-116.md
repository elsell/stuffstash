# Frozen candidate batch — native search placement

Not released. Production scope: M207, one assignment reordered in pinned
react-native-screens4.23.0 so placement/toolbar settings precede search-controller
attachment. Existing native search options and behavior remain unchanged.
The patch applies to all consumers, including Browse, Place, Settings, Expiration
and voice location selection. No color-picker or enlarged-text fixes are claimed.

Candidate evidence: baseline352224 reproduces bottom phone Settings search;
352366 accepts six other workflows on both devices;352442 accepts both filter
search workflows with reviewed command-clearance captures. Native test refinements
preserve query/selection/navigation assertions and fix demonstrated observation
problems. This is not yet acceptance of the installed production dependency.

Release gates: pinned dependency resolution; shared-header/search regression
checks; code critic; installed-patch eight-flow phone/iPad native acceptance;
required CI; TestFlight processing and changelog readback. Other comprehensive
audit findings remain tracked independently and do not expand this batch.

Proposed TestFlight notes: Fixed native search placement so search stays in the
header across inventory browsing, settings, and selection screens.
