# Android header and vector inspection

September16,2026: Pixel6 emulator, Android16/API36 revision7,1080×2400 at420dpi,
font_scale1.0, light appearance, on paul. Source archive db8d2bb8 with synthetic
fixture routes; no production service or user inventory was used.

Initial APK SHA-256:
`f28eed5587745193c4b11b65ed888e046373c708b1214b7e94c5a839ffbb54cc`.
The native build completed in6m30s; installation and cold launch succeeded.
[Home before correction](evidence/android-home-icons-before-db8d2bb8.png) shows
invisible Add, solid Profile, and unintended red dots beside both actions.

The pinned Expo UI55.0.17 VectorIconLoader parses pathData/fillColor and explicitly
ignores stroke properties and fillType. Thirteen shared vectors depended on
strokes. Its BadgedBox native implementation also supplies a default Badge when
the badge slot is absent. These explain the observed renderings; they are separate
from native hit-region and command behavior.

The candidate uses filled Material24px contours with pinned source mapping/license
beside the assets. It creates BadgedBox only for a positive count. The structural
regression failed first on an unsupported stroke; rejecting stroke/fillType and
accepting filled paths now pass. TypeScript, structural checks and10 focused
header/conversation/read-state behavior tests pass on paul. Critic found no blocker.

Patched APK SHA-256:
`1fde560e1a2f3a1a3751f1cf82514e38df6a4fac72ebae3319d09335c7f62e54`.
It includes the header/vector working-tree correction over db8d2bb8; incremental
build completed in53s. [Home after correction](evidence/android-home-icons-after.png)
shows recognizable Add, Notifications and Profile in order, with only the count2
notification badge. ADB interaction located the accessible icon labels, tapped
their centers, verified the Add/Profile destination markers and returned to Home
with system Back after each. These are production header handlers with fixture
destinations, not acceptance of the full Add/Profile workflows.

Other shared icon consumers, zero-count transition, disabled/removed actions,
TalkBack, dark appearance and broader Android journeys remain to be exercised.
The fixture omits production's StatusBar component; its light status-bar glyphs
are not evidence of a production status-bar regression. No TestFlight release is
claimed from this Android-only verification.
