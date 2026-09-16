# Android Add header commands

M226: the production-equivalent Add fixture requested a full-height form sheet with
a header. Android displayed the form without Close Add or Save item. Its native
hierarchy and [before capture](evidence/android-add-before.png) confirm the missing
commands at normal text size. Android form-sheet header support differs from iOS.

Production Add and its configured fixture now share `createAssetNativeSheetOptions`
Add options: Android native-stack card with header; iOS retains the full-height
sheet. This preserves Add's current header handlers, eligibility, retained draft
and busy-operation removal guard. The presentation test failed before the change;
18 focused presentation/dismissal tests, TypeScript and structural checks pass on
paul. Critic found no presentation-migration regression.

Rebuilt APK3a8ed4fb5553b4ba7b5c7deb5432bd4e6725059b3f5ea7c626a84ef5723e17d6
shows Close Add and Save item. Entering AuditAdd and tapping Save reaches the
synthetic command, whose deliberate failure displays Rejected draft: AuditAdd;
the name remains intact. [After rejection](evidence/android-add-after.png).
Tapping Close returns to the previous audit index. No real item was created.
This verifies visible actions, current input submission, failure retention and
ordinary return; it does not verify successful persistence, root-entry escape,
every optional Add field, or iOS regression acceptance.

Review identified M227 separately: production `src/app/add.tsx` uses only
`router.back()` for Close and has no fallback when opened without history.
The presentation fix does not resolve that existing navigation gap.
Conversation, inventory switching and checkout history also still request Android
form sheets with headers; inspect their tasks/actions individually before changing
presentation. Native Conversation inspection confirms its requested title is absent,
but its proposal Approve/Cancel remain visible, so it is not the same severity as
Add's missing save command.

Retained evidence: `/tmp/android-add-{header,after,rejected,closed}.xml`,
`/tmp/android-add-{red,green,structural,build}.log`. Build provenance is the isolated
db8d2bb8 fixture tree plus the previously recorded patches and this Add option
change, not a distribution build of the full current branch.
