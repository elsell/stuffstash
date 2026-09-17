# Phone fixture checkpoint: run 35042066124

Completed job 104625118262: **46/69 cases passed; 23 failed**. Source
d39e423321528376c41f3f78b4629e680cc0c66e, tested merge
627e5dae2647cadf19ad7012a010560915f082a8. The workflow has now completed with
failure: iPad fixtures passed 52/69 cases; phone onboarding passed and iPad
onboarding passed 2/3. These results belong to that source, not current head.
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
checkpoint. The completed iPad job is 104625118209; its log is retained at
`/tmp/native350420-ipad.log`.

Sharing's phone capture shows the keyboard covering the cancellation menu and
the invitation ellipsis accessibility bounds measuring only 21.7 by 6.7 points.
M193/M194 address submission keyboard dismissal and the native label's hit area.
See sharing-axis.md for evidence and pending verification.

The full/nested footer failures are isolated comparison fixtures, previously
documented in phone-sheet-comparison-350298.md. They are not new evidence that
the production direct-scroll filter sheet regressed; preserve that distinction.
