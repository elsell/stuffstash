# Full installed-patch native acceptance —35247151136

PR159 source1a45d0bdd5524fd1a79f59d6c0a396022f4aa602; tested merge
f7b3f995dd2ef0e9f26bba2f48d6dfd759075bbd. Full fixture run: phone105290144023
passes74/92, iPad105290143903 passes84/92. Exact outcomes are in
native-full-352471-results.csv; this is not an all-green audit.

All eight frozen-batch search journeys pass on both devices (16/16): ordinary and
preconfigured Place, managed/static placement, Settings, voice location, Browse
filters and Expiration filters. They verify full query, filtering, Clear/re-entry,
selection/navigation and command clearance where applicable. This run uses the
installed pinned patch, without the removed runner transformation. Reviewed
captures confirm header Search in phone Place and Settings, and iPad filter
commands above the keyboard accessory; retained with352471 suffixes in evidence/.

The separate focused run352470 still failed two observation checks. Its iPad
trace proves a true evaluation exceeded the waiter scheduling budget; that
failure remains recorded. This independent, already-running full run accepted
those same unchanged checks. The later eager-observation improvement is not in
this released source and has no native acceptance yet.

Onboarding phone105290143544 passes one executed test and skips two;
iPad105290143896 passes3/3. Do not call the skipped phone paths verified.
Required CI35247151149 passes, including native dependency resolution.

Remaining full-suite failures concern known color activation, text-entry and
normal/enlarged-layout checks. They remain comprehensive-audit work, not claims of
this search fix. Release scope stays M207; no assertion of overall UI conformity.
The original jobs were allowed to finish. A Bash observer slept120 seconds;
no observation timeout caused a restart. Selected artifact reads totaled under4MB.
