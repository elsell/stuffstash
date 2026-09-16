# iPad native results — run35095011021

Source403ef11c, merge093408914d8f3f229eb349eb91024bf08683ffdb, iPad mini(A17 Pro).
Final job104790103987 log records **73/83 passing,10 failing**, completing all
83 tests at13:45:32 UTC. Xcode exits65. GitHub's terminal job conclusion is
`cancelled`; retain both facts rather than deriving test completion from job status.
An earlier API log ended mid-snapshot with only45 passes/8 failures; it was partial.

Production and preconfigured Place search, managed/static search, Sharing recovery,
both notice geometry/command journeys and header-configured Add pass. These are
named journey results, not whole-platform acceptance or current-source verification.

Normal-size failures:

- Ordinary system color-picker opening fails. Independent target evidence from
  earlier runs is insufficient to replace this failed opening journey.
- Controlled address, ordinary controlled name and the no-accessory controlled
  name comparison lose or reorder characters.
- Default-assisted native field comparison fails keyboard readiness before typing.
- Move Here times out waiting for its candidate after retry.
- Voice location selection fails because clearing the query does not keep native
  search available (M232). Static search passes but does not prove this focused
  clear behavior.

Three enlarged-text journeys fail: Edit metadata, Edit tags and Move Here recovery.
Keep those queued behind normal-size remediation.

Log: `/tmp/native350950-ipad-final.log` locally. Artifact10448943886 (about1.4GB)
is downloading to `/tmp/native350950-ipad.zip` on paul. Inspect targeted attachments
before making visual or causal claims; do not extract the full xcresult twice.
Latest-source run35104157358 at14d7e06f has started independently through the queued
workflow. It includes subsequent photo, asset-menu and email fixes absent here.
