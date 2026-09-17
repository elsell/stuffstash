# Frozen follow-up batch after TestFlight 113.1

Production cutoff d8f4b4f0: M249 native keyboard-window boundary measurement
and M251 scoped iOS photo-viewer status-bar ownership, including the local pod
lock. Subsequent commits contain evidence and audit reconciliation only.

## Changed workflows and acceptance

- Browse and Expiration tag search: full query, tag selection, both footer
  commands completely above keyboard/accessory, overview return and application.
  Native351811 passes on phone and iPad; reviewed captures and frames are in
  native-boundary-351811.md. iOS pod compilation also succeeds.
- Shared photo viewer: valid visibility owns light status text over its black
  canvas; closing or removing the last photo restores the parent. Native351832
  passes all five Add-draft journeys on each device; reviewed light-appearance
  viewer and last-removal captures are in native-photo-status-351832.md.
- Source regressions: 1,919 tests across304 files, TypeScript and mobile structural
  checks pass on paul. Mounted cases cover measurement races/hide/detachment and
  viewer visibility/invalid selection/Android ownership. Critic review cleared
  the corrected keyboard-window coordinate contract. These do not substitute
  for native visual evidence.
- Integration: exact-head required CI must pass before merge. Preserve the
  source cutoff and verify release completion, Apple processing and notes readback.

The phone ordinary color-picker activation failure in351811 remains open; its
control is unchanged by this batch. Saved-photo, dark-appearance, swipe-dismissal,
rotation and multiwindow coverage remain in the comprehensive audit. This release
accepts the reported normal-text portrait corrections and does not certify all
adaptation states. No unrelated findings enter its release gate.

## TestFlight notes

- Fixed filter action buttons overlapping the keyboard on iPhone.
- Improved status-bar readability when viewing photos and returning to Add.
- Known issue: the custom color picker may occasionally fail to open; reopen the editor and try again, or use a preset color.

The version/build is assigned by the release workflow. TestFlight113.1 excludes
these corrections. This document records the next authorized batch, not a claim
that it has shipped.

## Released

Exact-head CI35185926029 passed; PR155 merged as
9c02cb48e4f4c42db053a98b582588c185f39335. Release35186356231 completed
successfully, including iOS upload105090973280 and changelog105093759613.
At2026-09-17T06:01:11Z the publisher verified TestFlight0.24.25(114.1)
notes after Apple processing. The frozen M249/M251 corrections are now released.
The sleeping Bash observer completed without model polling or job restarts.
Broader audit and adaptation findings remain open.
