# Phone fixture checkpoint: run 35042066124

Completed job 104625118262: **46/69 cases passed; 23 failed**. Source
d39e423321528376c41f3f78b4629e680cc0c66e, tested merge
627e5dae2647cadf19ad7012a010560915f082a8. The iPad fixture job was still running
when this report was written; this is not the whole workflow conclusion.
Evidence: completed job log `/tmp/native350420-phone.log`, retrieved through the
job-log API while the parent run remained active. Current branch fixes after
d39e4233 are not validated by this run.

Normal-size findings remain: Add/text-entry scenarios, color well target/picker,
Home action target size (36 versus asserted 44), full/nested footer layout,
invitation cancellation hit testing, and the Home return recovery selector.
The latter fails on multiple matching error elements; the subsequent selector
correction is not included in this run and is not claimed verified here.

Ordinary controlled input returned `Nve draft name`, ordinary single-line input
`Nive draft name`, and no-accessory input `Nve draft name`, each instead of
`Native draft name`. These are actual XCTest value mismatches, not proof of a
specific production cause. Pacing diagnostics added later are not in this run;
do not weaken the original full-value assertions or certify typing from unit tests.

The new voice proposal location journey successfully reached its retry/search/
selection steps and reopened the selected location. It then failed at source
line1552, `header.buttons["BackButton"].firstMatch.isHittable`, before Back was
tapped. Artifact inspection subsequently confirms a real missing Back control:
the visible chooser is presented as a sheet and its navigation bar contains only
the title and Search. See [captured screenshot](voice-location-missing-back-350420.png)
and [hierarchy](voice-location-missing-back-350420.txt). M192 adds an explicit native
Back; this is a product correction, not a selector-only relaxation.

Additional failures explicitly target enlarged text: asset region, command height,
detail commands, Edit metadata/tag and Move Here. Keep those recorded while
prioritizing the normal-size failures per the user's sequence. Expiration's
accessibility failure remains in the existing finding backlog as well.

No current-build native acceptance or TestFlight readiness follows from this
checkpoint. Continue the existing iPad job; preserve its authoritative handle.
