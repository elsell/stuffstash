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

The native-intent fix suppresses warm callback navigation while AuthSession retains
its original URL event and state/PKCE validation. Cold callbacks return to the root
without query parameters. Revised native sign-in, relaunch and isolation acceptance
remain pending; source tests alone do not close this batch.
