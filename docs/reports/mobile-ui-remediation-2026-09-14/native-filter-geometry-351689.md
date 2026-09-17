# Phone filter coordinate evidence —35168921441

Completed iPhone job105040747373 passes3/5; Browse keyboard clearance and
Expiration action hittability fail. Source966158e6 matches the released production
cutoff; this selection adds a fixture-only geometry probe. No new product fix.

Both Browse and Expiration retain the same probe: boundary x0,y812,width402;
keyboard x0,y495,width402,height379; calculated inset317. XCTest places the
probe (top140 in the sheet) at y202, consistent with the sheet origin62.
Show results/Apply is y429–483, Back y491–545, and Dismiss keyboard y485–529.
The measured sheet edge lacks the62-point presentation origin while keyboard
coordinates include it. This supports the coordinate-space mismatch rather than
a missing settled event. The independent probe is not a synchronized trace of
every production hook callback.

[Retained Browse hierarchy](phone-filter-geometry-351689.txt). Original phone
artifact10476902860 remains on paul at /tmp/native351689-filters-phone.zip; log
is /tmp/native351689-filters-phone.log. iPad artifact10476988817 has not yet been
reviewed in this follow-up.

Next correction must measure the sheet boundary and keyboard in the same window
space, or use native keyboard layout ownership. Do not hard-code62: iPad sheet
position, rotation and resizing differ. Verify actual Browse/Expiration query,
selection, keyboard accessory clearance and full Apply/Back reachability on both
devices. M249 remains unresolved after the explicitly authorized113.1 release.

## Follow-up candidate and producer coordinate correction

Pinned RN0.83.6 RCTKeyboardObserver converts notification screen frames into
RCTKeyWindow before publishing KeyboardMetrics. The earlier screen-space wording
was misleading: the boundary must match that window coordinate space. The local
Expo view now converts its bounds into its own window only when that window equals
RCTKeyWindow; detached/other-window views return unavailable. A typed async port
keeps the existing generation/hide/settled behavior and clears unavailable or
rejected measurements. Android's separate adapter is unchanged.

The three port tests failed before implementation; all1,919 mobile tests across
304 files, TypeScript and structural checks then passed on paul. The native module
is discovered by pinned Expo autolinking. Code critic caught the initial screen
conversion mismatch and cleared the corrected window-identity conversion. Pod
lock resolution, Swift compilation and native phone/iPad acceptance remain pending.
The original probe remains historical evidence of the Fabric path, not a measure
of the new production adapter. M249 stays open.

macOS CI35180945960 resolves the local native module with pinned CocoaPods1.17.0.
The retained lock adds only StuffStashSheetBoundary1.0.0, its local path, existing
ExpoModulesCore/React-Core edges and the generated podspec checksum; no existing
dependency version changes. Swift compilation and native interaction verification
remain separate requirements.
