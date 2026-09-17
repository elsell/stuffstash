# Search-controller attachment experiment —35219303504

Sourceff8277757e93b2c2316201a3f7fd4fd15dcf3474. A runner-only transformation
moves UINavigationItem.searchController assignment after preferred placement and
toolbar-integration settings in pinned react-native-screens4.23.0. It leaves
conditional compilation and property values unchanged and rejects reapplication.
The production dependency patch and locks do not contain this experiment.

Phone105195236910 passes all7 selected workflows, including previously failing
preconfigured Place, production Place, managed/static placement, Settings,
Browse and Expiration search. iPad105195236694 passes6/7; Expiration fails at
interactive-keyboard readiness before typing. This is not an all-green result
or evidence that iPad Expiration is accepted. Exact results are retained in
native-search-order-352193-results.csv. Captures still need visual review.

This is promising candidate evidence for M207, not causation proof or production
acceptance. Earlier unmodified builds sometimes passed production search. Before
promoting any patch, inspect phone preconfigured Place and settings geometry and
the failed iPad keyboard capture. Keep the existing integration assertions and
shared search consumers in scope; do not relax keyboard acceptance to ship this
experiment. A same-source unmodified baseline would strengthen attribution.

Ten fixture installer checks and code critic passed. An in-memory transformation
check on paul verified exactly one moved line and unchanged line multiset, with
reapplication rejected. The Bash observer slept120 between checks and did not
restart jobs. Both complete artifacts remain available from this run.
