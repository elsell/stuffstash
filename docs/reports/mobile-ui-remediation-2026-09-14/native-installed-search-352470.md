# Installed search patch —35247098774

Source1a45d0bdd5524fd1a79f59d6c0a396022f4aa602. Both devices pass7/8 installed-patch
journeys; Browse fails keyboard readiness on iPad and command-clearance observation
on phone. No runner-side transformation is present. All other seven workflows,
including native placement, Settings, Place, voice and Expiration pass both.

The iPad trace proves the readiness predicate returned true after4.0277 seconds,
but the wait reports timeout because evaluation started1.0397 seconds into its
five-second window. Total5.0679 seconds. Later diagnostic confirms a hittable t key.
Phone diagnostics show both buttons hittable and contained, Back bottom483 and
Dismiss keyboard top485. These are subsequent state, not a passing timed assertion.
Exact traces are retained in evidence/. This is concrete evidence to remove the
waiter's initial scheduling delay, not increase the allowed deadline.

Evaluate immediately and use only the remaining five-second budget for subsequent
waiting. Reject any positive evaluation that finishes after the deadline. Preserve
all keyboard/query/geometry/hittability conditions. Revalidation remains required.
The full PR native run35247151136 was already running and was not restarted.
