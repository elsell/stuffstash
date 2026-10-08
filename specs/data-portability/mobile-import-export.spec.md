# Mobile import and export workspace

## Approved task and pattern — October 8, 2026

Inventory Settings exposes one **Import and export** navigation row, replacing
separate immediate-export menus and archive actions. Open a normal native stack
screen with standard back navigation, not a modal sheet. The household restore
entry opens the same workspace in import mode, including households without an
inventory. Preserve the existing native system file picker and share sheet.
This is a persistent task destination: configuration and server job history earn
a full screen. It follows the platform interaction review standard and Apple's
HIG distinction between a bounded sheet task and a navigable destination.

## Actions and hierarchy

Present Export and Import as peer modes with a native in-place choice. With an
inventory, default to Export and backup archive; without one, offer Import.
Export formats are Backup archive, JSON data and CSV spreadsheet, with short
explanations: archive preserves inventory and optionally photo/file bytes; JSON
preserves inventory records/referenced definitions without attachment bytes; CSV
is a spreadsheet report, not a restorable backup. Archive photo/file switches
start enabled and excluding either clearly identifies a partial backup. One
primary action submits the selected format. No nested menu that starts an export
merely by choosing a format.

Import uses a Stuff Stash ZIP archive and the existing upload, validation, explicit
review/name and approval flow to create a new inventory. Existing inventories are
not overwritten. Show the existing preview counts, omitted attachments and key
remappings before approval. Keep review within the normal navigation experience,
with a clear return to activity and no second sheet.

## Persistent activity and accurate assurance

A shared activity section stays available beneath task configuration, across both
modes. Discover the signed-in user's accessible export and restore jobs across the
current household with the existing household-scoped API. Label this scope clearly;
never imply all entries belong to the currently selected inventory. Show job kind,
creation date, status, availability/expiry when relevant, and existing download,
review, open, retry or cancel actions. Put queued/running/awaiting-review jobs first,
then terminal jobs, newest first within each group. Preserve pagination, deduplicate
by job ID, refresh on focus/foreground, and poll only while work is active.

Accepted archive jobs live on the server and survive navigation or app restart.
Explain that users can leave and return once a job is accepted. A local archive
upload, file download/share or direct JSON/CSV export must instead say to remain
on the screen until that transfer finishes; leaving cancels local transfer only.
Do not promise resumable uploads or durable history for immediate JSON/CSV exports:
they remain the existing direct downloads, explicitly distinguished from saved
archive activity. Do not replay a submission automatically on return or retry an
uncertain write with a new idempotency key. Failed jobs remain visible for recovery.

Scope/account changes cancel local work and clear old data. Permission errors
clear inaccessible jobs/review; a transient refresh failure retains readable
history with a retry notice. Reuse existing authorized repository paths, request
keys, cancellation and archive security boundaries; no new API or persistence
contract is introduced.

## Acceptance and delivery

Critical automated checks cover unified settings entry/normal navigation, correct
format dispatch, mixed job ordering, leave/reopen discovery without replay, and
restore review before approval. Retain existing transfer/auth/scope safety tests;
avoid exhaustive prop snapshots. Verify iPhone layout, scrolling to the final job,
selection, history return and upload/accepted-job wording in the existing native
fixture workflow where available. Native evidence remains distinct from source
checks; user-device checks go on the consolidated checklist without blocking
otherwise verified delivery. Update public export guidance and release notes.
