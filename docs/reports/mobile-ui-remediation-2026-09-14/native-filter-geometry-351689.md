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
