# Android filter sheet inspection (in progress)

Android16 Pixel6 emulator on paul, normal text, light appearance. The initial
Browse filter sheet at detent0.7 lacks its title and both footer commands;
expanding to1.0 reveals Show results/Cancel. The ADB assertion that Show results
exists fails at initial presentation. [Before](evidence/android-filters-initial-before.png).

Pinned native-stack documents Android form-sheet headers as unsupported and
provides `unstable_sheetFooter` because JS-only footer placement is problematic.
The Android-specific candidate uses that native footer, measured scroll clearance,
an accessible body title, stable presentation and current committed callbacks.
iOS retains its existing direct-scroll layout.
[Guidance](https://reactnavigation.org/docs/native-stack-navigator/).

The first patched warm sheet shows title and both actions at initial0.7:
[capture](evidence/android-filters-native-footer.png). That is rendering evidence,
not full apply/cancel, last-row or keyboard acceptance.
Warm candidate APK SHA-256:
`3ca19be05b5bcd0016650fd96b4beddfef63d4f760ef22c6b5dcdefb96d6fd9a`.
Later root-fallback APK SHA-256:
`7fcf72fc844f3dcc777beeb0123ddaf473c898927705a17b668e29ca4459b942`.
Both are db8d2bb8 synthetic archives with the reviewed icon fix and the then-current
uncommitted filter adapter; neither includes the latest top-inset correction.

Cold-link verification exposed a first-route distinction: Android renders a root
screen rather than a sheet, so the native sheet footer disappears. A full-screen
footer fallback now renders the actions, but its inspected title overlapped the
system area. Top inset reservation has been added and needs a new native build.
Code review also caught competing headerShown=true assignments in the Expiration
route/page; these were removed so the adapter owns Android visibility and iOS uses
stack registration. Native re-verification of those latest changes is pending.

Fourteen focused tests, TypeScript and structural checks passed on paul before the
latest top-inset adjustment. Root footer expectations failed before the fallback;
the first test attempt also exposed a missing useRouter export in the navigation
fake, subsequently corrected. Review has not yet accepted the complete final pass.

Remaining checks: cold-root top inset; root Cancel escape (current production
routes call router.back unconditionally); initial/expanded detents; last tag row;
nested-page search without native sheet headers; draft apply/cancel; keyboard;
Expiration parent rerenders; Android dark appearance and assistive technology.

## Latest native pass

The current APK (`0789c20cf7ab484362c7b168355c0d851a0bd07295e184c729aca5950de06dbc`)
adds top inset, consistent Android header ownership, and root-entry Cancel recovery.
It remains the db8d2bb8 synthetic archive plus icon and filter patches, not a release.
Cold Filters title starts at y128; both actions are visible, and tapping Cancel
returns to the audit index. [Root capture](evidence/android-filter-root-after.png).
Warm initial detent shows title/actions. Scrolling Tags expands the sheet; the last
row is above the footer, can be selected, and Show results returns the confirmed
fixture result `Browse selected tags: audit-last`.

All 1,873 mobile tests in 290 files pass on paul, as do TypeScript and structural
checks. Final critic review found no further source blocker. These checks do not
close full filter acceptance: native Tags search is absent in the headerless Android
sheet (M221). Expiration, keyboard, dark appearance and assistive technology remain.
The white status glyphs in the synthetic root capture are not production evidence:
the fixture omits production StatusBar configuration.

## Keyboard rejection of the footer candidate

Adding in-body search reveals a native crash on focus: RNScreens
`ScreenFooter.onParentLayout` calls `sheetTopInStableState` while the bottom sheet
is settling during keyboard inset changes. AndroidRuntime reports
`IllegalArgumentException: [RNScreens] use of stable-state method for unstable state`.
The process exits to the launcher. The pinned implementation confirms this call
has no dragging/settling branch. The initial/last-row successes above do not
establish keyboard safety, and this candidate must not ship as accepted.

The in-body search has a failing-before/passing-after interaction test; critic also
caught stale page events, now guarded with a keyed input, committed handler ref,
focus cleanup and unmount cleanup. Tests verify hidden events are ignored, current
input recovers on focus, and removed-page events cannot restore old queries.
These guards are not a fix for the native footer crash. The next candidate is an
Android full-screen native stack presentation with a layout-owned footer and
keyboard handling, preserving iOS sheets. It requires spec and native acceptance.

## Replacement presentation and keyboard checks

The replacement uses Android native-stack `card` presentation, a normal title bar,
in-body search and a flex-layout footer. It removes `unstable_sheetFooter` entirely.
Initial native typing no longer crashed, but Gboard covered the footer; an explicit
KeyboardAvoidingView with the native header offset corrected that.

APK SHA-256 `36fe660b24f35a8b9b7d8b2d98f4feac9d03495ae95392807fed71eac5adbe92`
(db8d2bb8 fixture archive plus current icon/filter patches) was tested at normal
size on the same Android16 emulator. Browse and Expiration Tags both accept
`Tools`, narrow to the matching row, and keep primary/Back controls above Gboard.
Browse Back removes search and restores the filter overview. Expiration selection
and Apply return to the prior Browse filter route in the fixture stack. [Keyboard capture](evidence/android-filter-keyboard-after.png).
This is synthetic interaction evidence, not production backend acceptance.

The final full source suite before the KeyboardAvoidingView adjustment passed
1,876 tests in291 files; the final wrapper passed the3 focused filter tests,
TypeScript and structural checks on paul. Broader search/clear/last-row retesting,
dark mode, TalkBack and production navigation remain open; iOS search placement
must still be verified. The earlier native-footer candidate is superseded, not an
accepted fallback.
