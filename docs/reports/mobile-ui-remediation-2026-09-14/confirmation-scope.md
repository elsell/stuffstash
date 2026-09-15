# Native dialog and confirmation scope

Initially inventoried at a91953a1 and line references refreshed with M135, `confirmation-call-sites.csv` records24 static alert/dialog occurrences
across production mobile TypeScript:23 callers plus the shared Alert implementation.
This includes informational errors, destructive confirmations, draft discard and
photo-source selection. They are not24 equivalent destructive workflows and are
not24 newly certified surfaces. The inventory nominates sites for review; its
individual rows remain pending unless the review column links subsequent findings. Native acceptance is tracked separately.

S129 was anchored only to AppFeedback.tsx. Source inspection shows that
`showDialog` supplies ordinary/cancel native alert buttons and is used by session
expiry, push-open failure and photo-removal failure. Destructive workflows instead
invoke Alert directly. Consequently the shared wrapper cannot establish correct
confirmation wording, cancellation, one-shot use or visit ownership everywhere.

Earlier findings already cover some callers (asset lifecycle, provider discard,
Sharing cancellation and customization). Preserve their existing evidence rather
than treating this inventory as a new failure in all of them. The remaining review
must inspect each call's initiating state, commit/cancel actions, pending guard,
resource/account identity, late callback behavior and consequences. Native
phone/iPad screenshots and actual Cancel/Confirm journeys remain separate gates.

The system-native Alert primitive is appropriate for many irreversible choices;
this does not mean every informational error or source choice needs an alert.
Review task fit at each site before introducing a shared abstraction. No universal
migration is proposed here.

This inventory covers literal Alert.alert/showDialog call expressions in .ts/.tsx.
It is a source-review index, not an AST proof that no other modal mechanism exists;
menus, sheets, navigation guards and platform-owned permission prompts have their
own surfaces in the main matrix. No matrix cell is promoted by this inventory alone.
