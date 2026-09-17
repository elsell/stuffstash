# iPad native audit — run35156794515

Job105001197554 tests source2043abb0 through merge2209ff1c347187e92eef215f965b5fe68e040df2
on iPad mini (A17 Pro). It completes80/87 cases: all55 required journeys,20/24
diagnostics and5/8 enlarged-text follow-ups pass. The four isolated text-entry
comparisons and three enlarged-text failures remain in the full audit. These
counts are native assertion evidence, not a claim that every capture was reviewed.

## Selected visual review

- [Browse tag search](evidence/ipad-browse-keyboard-351567.png): the selected Tools
  row and complete Show results/Back actions remain above the keyboard accessory.
- [Expiration tag search](evidence/ipad-expiration-keyboard-351567.png): Apply
  filters and Back are fully visible above the accessory while search is focused.
- [Settings header](evidence/ipad-settings-header-351567.png): Add and Search
  appear as separate native header buttons. This build still uses the inaccurate
  audit-customization fixture title; the corrected Tags title needs its rerun.
- [Sharing recovery](evidence/ipad-sharing-recovery-351567.png): the complete
  audit@example.invalid invitation, one-time link, Copy/Share and inline simulated
  sharing failure remain visible. The header Back control is unobstructed, and
  the pending and cancelled invitations are distinguishable. The fake deliberately
  does not open an external destination; this is not evidence of external sharing.

All four captures use normal text/light appearance. This build predates the new
iOS email adapter223d6d0a, so its sharing pass cannot accept that correction. The
phone keyboard overlap remains a batch blocker despite iPad passing.

Artifact10474646442 is retained on paul at `/tmp/native351567-full-ipad.zip`
(1,459,930,308 bytes); only selected captures were copied locally. Complete log:
`/tmp/native351567-full-ipad.log`.
