# iPad onboarding run35054957505

Job104663309325, source8a256a2a, completed with2 of3 tests passing. Source includes
the native onboarding command migration. The connection/keyboard journey failed
the exact address assertion: native field value was `h` instead of the complete
synthetic HTTPS address. Inside-form keyboard dismissal and landscape passed.

Initial evidence was the terminal XCTest log, OnboardingAuditTests.swift line89.
Subsequent inspection of artifact10430906138 changes the diagnosis: the recording
and final screenshot show the complete address. The field stays focused, first
showing `h`, then the full string; the keyboard stays visible. Thus permanent text
loss is not established for this case. The immediate assertion observed a partial
value before completion. Exact reasons for the delay remain unproven.

Retained [initial screenshot](ipad-address-initial-350549.png) and
[final screenshot](ipad-address-complete-350549.png) preserve both states. Recording
frames62–88s show complete text by approximately71s, retained through teardown.
The log records typeText synthesis at66.89s, initial capture70.6s and final
capture87.23s. The separate inside-column case passes the same full-string check.

The candidate adds a five-second exact-value predicate before capture and the
original equality assertion. It neither retypes nor slows injection, accepts no
substring, and treats timeout as failure. Structural checks pass remotely; native
compilation/rerun remain pending. Do not extrapolate this observation to differently
reordered React Native fields or call address entry fixed from artifacts alone.
Earlier onboarding passes do not certify this newer source. Other jobs in this
run remain active and must not be restarted.

## Phone onboarding result

Job104663309509 failed its applicable connection/help journey; two iPad-only cases
were skipped. Address equality passed. The five-second keyboard-disappearance
predicate after the margin drag timed out. Artifact10430568792's
[final screenshot](phone-onboarding-dismissed-350549.png) and hierarchy show the
keyboard absent with the full address retained. This later state does not convert
the deadline failure into a pass.

The log spends over a minute repeatedly resolving scroll bounds before the drag
(67.91s,93.63s,125.14s; gesture139.51s). The candidate reads each required frame
once per gesture and retains the same coordinates, containment assertion,
keyboard deadline and subsequent checks. Structural check and critic review pass;
native rerun remains pending. No keyboard-dismissal production fix is inferred.

Focused text-entry run35056372549 is independently active atfdbf30bf, jobs
104667325335(phone) and104667325488(iPad). It includes the default native assisted
comparison but not onboarding; the full run350549 fixture jobs remain untouched.
