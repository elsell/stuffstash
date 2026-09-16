# Android Conversation header

M230: Android form-sheet presentation omitted Conversation's title, Close voice
session and New conversation controls. Proposal Approve/Cancel remained visible.
Android now uses a native-stack card; iOS retains0.42/0.88 detents and its larger
initial detent. Production and fixture share these options.

The presentation test failed before the change;3 focused presentation/header tests,
TypeScript and structural checks pass on paul. Code critic found no source blocker.
APK a3ec1d58c1f56909c8c05079d9860a269d814c42d30c1ddb022755cdfe5932be
shows [both header commands](evidence/android-conversation-header.png), context
below the header and visible proposal actions at normal text size.

Native New conversation opens its confirmation. Keep conversation preserves the
proposal. Close returns to the audit index; reopening retains the proposed drill.
Opening containing location shows the fixture's initial lookup failure; Retry loads
Garage bin, and selecting it returns to Conversation with the edited location and
Approve still present. These checks use synthetic ports and do not record audio,
contact a model provider or mutate a real inventory.

An additional confirmed-reset observation did not reach an empty conversation:
the synthetic proposal reappeared with its original Inventory root location.
Source inspection shows the fixture's route-local seeding refs can reset across
navigation and seed again after the controller clears its plan. Treat this as an
unresolved fixture-isolation issue, not evidence that production reset failed or
that empty-conversation acceptance passed. That scenario needs provider-lifetime
seeding before it can establish reset behavior. Existing source confirmation/reset
tests remain passing.

Retained native hierarchies: `/tmp/android-voice-{dark,header,new-confirm,closed,retained,location,location-retry,location-selected,reset}.xml`.
Logs: `/tmp/android-voice-header-{red,green,structural,build}.log`.
iOS native acceptance, physical audio and individual composer-control appearance
remain open. This fixture APK is not a TestFlight/release artifact.

## Reset fixture follow-up

The seed now activates only on first proposal-route entry and remains mounted
at provider lifetime. A review caught eager provider seeding before entry; that
would contaminate Home, so activation remains route-triggered. Production
conversation behavior is unchanged.

Android APK `cdb7654eca5c5a1bd3ca18e5eed688154b2ccc1d0f9fbdff8b52cca3ab179d2e`
now passes Close/reopen retention and confirmed New: the proposal disappears and
[the empty composer and introductory text](evidence/android-conversation-empty.png)
are present. A fresh launch into Home in tabs exposes “Start voice interaction,”
confirming no proposal is seeded before Conversation entry. Same API36 Pixel6
runtime at normal text size; synthetic ports only.

Code critic accepted the corrected lifetime/activation. TypeScript passes in the
current source validation tree on paul. An initial check in the selectively
patched native build tree failed on stale unrelated test files; it is not used
as source validation. Native build and installation passed.
Evidence: `/tmp/voice-seed-{retained,confirm,empty,idle-home}.xml`;
logs on paul: `/tmp/android-voice-seed-build.log`,
`/tmp/voice-seed-source-check.log`. Remaining iOS/audio coverage is unchanged.
