# Android native-control appearance

M225: with Android16's device theme light and the app explicitly dark, the native
Cancel command retained dark text. The enabled label measured1.97:1 against the
app background at normal text size. [Before](evidence/android-compose-dark-before.png).
The pinned Expo Compose Host defaults to the system theme independently of the
React appearance context (`HostView.kt`).

A shared NativeComposeHost now supplies the resolved app appearance to all seven
Compose adapters: commands, sheet actions, header actions, action menus, refinement
buttons, read-state buttons and conversation buttons. Existing sizing, disabled
states, handlers and Material colors are preserved. This prevents each independent
Compose root from making a different theme choice.

On rebuilt fixture APK fa51a6026a8c62648792f076adcd5d30be280dd9246539e406f106877023f0cf,
the same enabled Cancel label measures10.89:1 in dark and8.96:1 after switching back
to light without remounting. [After](evidence/android-compose-dark-after.png).
Measurements use the two dominant interior label/background colors within the
observed Cancel text bounds (486,2222)–(596,2265); they do not certify every color
or control. The baseline contrast assertion failed before the source change and
passes after it. Disabled Move ignores a tap; selecting a destination then tapping
the committed enabled command records Move received. An immediate tap before
selection settled did not execute, so that unsuccessful observation is retained
separately from the settled-state assertion.

TypeScript, the mobile structural check and Android release-variant build pass on
paul. Code critic found no confirmed blocker. This is an isolated debug-signed
fixture build, not a distribution artifact. Other six adapter families share the
source correction but still need individual native appearance checks. The white
navigation header in this fixture is separate: its root Stack sets tint but no
header background, unlike the production Android tab-header adapter. No production
header-theme acceptance follows from this screenshot.

Evidence: `/tmp/android-appearance-{red,green}.log`,
`/tmp/check-android-cancel-contrast.py`, `/tmp/android-footer-disabled.xml`,
`/tmp/android-footer-enabled.xml`, `/tmp/android-footer-enabled-settled.xml`,
`/tmp/android-appearance-{check,structural,build}.log`.

## Additional normal-size dark appearance checks

On the same fa51a602 APK, the [Home header](evidence/android-home-dark.png) shows
Add, Notifications and Profile in order, with only the notification count badge.
A native tap increments the fixture's notification activation count from0 to1.
This verifies rendering plus the notification handler in dark appearance; it does
not repeat every header navigation probe.

The [Sharing menu](evidence/android-sharing-dark-menu.png) renders the native
Cancel invitation action legibly in dark appearance. The Create Invitation command
reaches the synthetic repository, whose first response deliberately lacks a link.
The resulting pending invitation remains available to cancel. The first confirmed
cancellation deliberately fails and displays its item-scoped error; reopening the
menu and confirming again reaches Cancelled, removes that error, and preserves the
email field. No real invitation is sent. This exercises command and action-menu
adapter families in addition to the previously verified sheet actions and header.
Read-state, refinement and conversation button families still lack this native
appearance sample. Android success does not resolve the separate iPhone Sharing
keyboard/cancellation finding.

The production root (`src/app/_layout.tsx`) explicitly supplies
`headerStyle.backgroundColor = palette.surface`, unlike this fixture root.
The white fixture header is therefore not evidence of a production root regression.
Retained hierarchies: `/tmp/android-home-dark{,-tap}.xml`,
`/tmp/android-sharing-{dark,created,menu,confirm,error,retry-menu,cancelled}.xml`.
