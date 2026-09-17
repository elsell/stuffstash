# Native run35034075257 — terminal evidence

Sourceaf4659f4557f9421afa32c38e9b4bf94dadf0978, before M166 and later fixes.
Both onboarding jobs passed. Phone fixture job104600926137 finished45/64 passing;
iPad fixture job104600925964 finished52/64 passing. These are XCTest log results,
not newly inspected screenshots or acceptance of current-head changes.

Phone normal-size failures include Add typing/readiness, color-well AX size,
Home action AX size, controlled address input, controlled/uncontrolled ordinary
input and uncontrolled input without accessory, full/nested diagnostic footer
layouts, and non-hittable invitation cancellation. Enlarged-text failures remain
tracked separately. AX frame height does not alone establish native touch area;
diagnostic footer failures do not establish production filter failure.

iPad normal-size failures include Add typing/readiness, color-well and Home AX
size, controlled address/ordinary input, place-content search and settings
collection search. The settings case reached Add successfully, then failed while
waiting for the native search field after tapping Search. Phone passed that case.
Do not infer its cause without inspecting native state/artifacts. Three iPad
failures concern enlarged-text metadata/tag/move recovery.

Character-loss and native target findings remain unresolved; passing siblings do
not close them. This build does not verify the newer return-error reveal, native
editor migrations, voice callback/capture ownership or proposal selection route.
Logs: `/tmp/native350340-phone.log`, `/tmp/native350340-ipad.log`.
