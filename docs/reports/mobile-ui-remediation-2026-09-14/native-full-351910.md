# Full native audit —35191010731

PR headf4c00a408c78641d208187dd622414e050588006, tested merge revision
620c9d0648ff496303cd6d238627996a8ee1a9b6. Phone job105103362498 passes68/88;
iPad105103362541 passes78/88. Both standalone onboarding jobs succeed (phone
one passed/two skipped, iPad three passed). Assertions are not blanket visual
acceptance. Full artifacts are retained on paul as /tmp/native351910-phone.zip
and /tmp/native351910-ipad.zip; selected RGB evidence is committed.

The added native RGB editing test passes on both with unchanged Green/Blue
channels through three controlled edits; see native-color-rounding-351887.md.
Browse/Expiration keyboard clearance, Add photo paging/removal and photo-removal
failure recovery pass on both. Account command/cancel/recovery also passes on
both, resolving the earlier run's missing destination coverage for this sample.

Against the historical55 required checks (which do not include the new RGB
case), phone passes53 and iPad52. Phone failures are ordinary color opening and
Expiration accessibility-text clipping. iPad failures are ordinary color opening,
notification delivered-touch completion wait, and full-fixture onboarding keyboard
readiness. Standalone onboarding success does not erase the different fixture
journey failure. Diagnostic text-entry/placement and enlarged-text cases remain
separate from the user's normal-text-first remediation priority.

The full case outcomes are retained in native-full-351910-results.csv. Inspect
failure artifacts before attributing the untriaged failures to product defects.
The existing phone preconfigured Place search comparison still fails; the iPad
comparison passes. Ordinary color opening fails on both. No repeated run was
started to replace either result.
