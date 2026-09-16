# iPad onboarding run35054957505

Job104663309325, source8a256a2a, completed with2 of3 tests passing. Source includes
the native onboarding command migration. The connection/keyboard journey failed
the exact address assertion: native field value was `h` instead of the complete
synthetic HTTPS address. Inside-form keyboard dismissal and landscape passed.

Evidence: terminal XCTest log, OnboardingAuditTests.swift line89. Screenshots and
recording are not yet inspected. This does not establish whether keyboard focus,
SwiftUI host updates, automation cadence or text synchronization caused the loss.
The production address adapter is already a native SwiftUI TextField; attributing
all text loss to React Native controlled inputs would contradict this result.
Retain the exact-string assertion and inspect focus/keyboard events before changing
input architecture or test cadence. Earlier onboarding passes do not certify this
newer source. Other jobs in this run remain active and must not be restarted.
