# iPad run35050407693

Job104649681823 finished with58 of72 scenarios passing,14 failing. Source802e4955,
test merge1eadad876d055bc3db688a0ef5e9d62842be77b2; later fixes are excluded.
Evidence here is the retained XCTest log, not fresh screenshot inspection.

Failures: Add draft navigation/rejected-save(2), color picker(2), controlled input
without accessory and ordinary controlled typing(2), Edit metadata/tags at large
text(2), Home header frame, Move-here at large text, Place search, provider
credential keyboard activation, settings search, and voice location return.
This preserves failures rather than weakening assertions.

Both Home Return cases pass. The notification center/edge/corner hit probe passes;
the older header frame assertion still fails and is not proof of a small effective
hit target. Photo removal failure presentation passes. Sharing's complete recovery
journey passes on iPad while phone fails after the menu reopens the keyboard.
Place and voice-location cases pass on phone but fail here; do not generalize one
device's result to the other. Native investigation remains open.
