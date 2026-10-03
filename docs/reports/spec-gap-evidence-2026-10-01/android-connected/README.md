# Connected Android acceptance

The initial normal-app build at `b3d0fb6d27c945ceb96e816746483db145b20395`
passed [hosted packaging](https://github.com/elsell/stuffstash/actions/runs/37107424166).
The APK checksum and runtime outcomes are in `initial-runtime-result.json`.

The API36 emulator used the isolated Paul API/Dex/SpiceDB stack with synthetic
accounts. No fixture routes, injected sessions or intercepted requests were used.
The exact test CA was added to the disposable emulator's system trust store and
through Android Settings for Chrome. The app's TLS configuration was unchanged.
Android's [Conscrypt trust stores](https://source.android.com/docs/core/ota/modular-system/conscrypt)
are separate from app code. This does not prove trust configuration on a physical
user device. The emulator hostname was pinned to the server's IPv4 address after
observing an unreachable IPv6 link-local connection; TLS hostname checks remained.

Real browser sign-in returned to the app but displayed **Unmatched Route** for the
app-owned callback. Authorization-code query parameters and authentication screens
are intentionally absent from retained evidence. Force-stop and launcher restart
then displayed the authenticated inventory and its seeded assets, as shown in
`owner-home-after-relaunch.png`: session persistence worked despite the return
navigation defect. Cross-principal isolation was not attempted in this build.

## Revised runtime acceptance

The normal release-mode APK at `43dabdcc6c29a6f421a446229830c4a8ff9cea54`
passed [hosted packaging](https://github.com/elsell/stuffstash/actions/runs/37108659742).
Its checksum and scoped results are in `fixed-runtime-result.json`.

Real browser sign-in now returns directly to Home with the seeded inventory and
assets (`owner-home-fixed.png`). Force-stop and launcher restart preserved that
session. Native Account → Sign Out → confirmation returned to Connect with the
server retained. Signing in through Dex as the separate synthetic account showed
household onboarding, without the previous account's inventory or asset data
(`second-account-onboarding.png`).

After creating that account's own household, explicitly navigating to the app root
showed its empty Home Inventory (`second-account-home.png`). Opening the first
account's asset URL in this authenticated session showed “Asset unavailable” and
no owner asset content (`second-account-denied-owner-asset.png`). These checks
cover this representative account switch and foreign-asset read, not every role
or native workflow. Existing API authorization tests remain separate evidence.

**Follow-up discovered:** household creation succeeded but initially landed on
“Containing location” with “This proposal is no longer available for editing”
(`household-created-wrong-route.png`). Root navigation recovered Home. The full onboarding journey did not pass in that build. The subsequent
[onboarding acceptance](../android-onboarding-archive/README.md) records the fix
and fresh-household runtime verification; no unchanged APK retry was used.

The native-intent fix suppresses warm callback navigation while AuthSession retains
its original URL event and state/PKCE validation. Cold callbacks return to root
without query parameters; cold callback behavior has source-test coverage only.
No iPhone, physical-device, TalkBack or authenticated archive acceptance is implied.
