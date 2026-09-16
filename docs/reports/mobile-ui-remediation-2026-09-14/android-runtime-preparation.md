# Android runtime preparation on paul

Initial September16 inspection: ADB was installed, but no devices were connected. No emulator or
sdkmanager was found on PATH or in the inspected common SDK locations. Java21.0.9
is available, and john has read/write ACL access to `/dev/kvm`. The home filesystem
has17GiB free; `/tmp` is a memory-backed filesystem and is not the SDK destination.

Initially installed the command-line tools under
`~/.cache/stuffstash-android-audit/sdk/cmdline-tools/15859902` on paul. Google's
numbered archive passed its published SHA-256 check before extraction/execution:
`4e4c464f145a7512b57d088ac6c278c03c9eea610886b35a5e0804e74eedf583`.
Source: [Android tools download](https://developer.android.com/studio).
The package listing succeeds; this tools version warns that sdkmanager is
deprecated in favor of its bundled Android CLI.

Google repository metadata identifies these stable packages. Emulator and system
image, platform and build tools are now installed. Repository XML is
retained in the same remote audit directory.

| Package | Revision | Archive | Published SHA-1 |
| --- | --- | --- | --- |
| Emulator |37.1.11|emulator-linux_x64-15917651.zip|1b1f78891abf8ec268264356e1365c25519e8379|
| API36 Google APIs x86_64 image |7|x86_64-36_r07.zip|c6bf44bdcd885bb902b4ba752d111a073ad7a817|
| API36 platform |2|platform-36_r02.zip|2c1a80dd4d9f7d0e6dd336ec603d9b5c55a6f576|
| Build tools |36.0.0|build-tools_r36_linux.zip|b0b6376977657e8ad9b969bacf4093601da2c6fb|

Metadata: [SDK repository](https://dl.google.com/android/repository/repository2-3.xml),
[Google APIs images](https://dl.google.com/android/repository/sys-img/google_apis/sys-img2-3.xml).
Downloads total about2.36GB compressed; expanded images, virtual-device data,
native toolchains and build caches require additional capacity. Check that capacity
before provisioning. Verify numbered downloads against the recorded checksums and
record SHA-256 digests of acquired artifacts. Do not substitute the preview
emulator37.2.9 also listed in the metadata.

## Boot evidence

Emulator37.1.11 passes its KVM acceleration check. The API36 revision7 image passed
its published checksum. Acquired archive SHA-256 digests:

- Emulator: `95771e0ae431897b2a4bd2d97fa095f29a8b0624a7b216baf529f9306161c266`.
- System image: `b1bb0769d0bed7698e61f203d7dc9bf6e7c37cd01a39d0d8788a11186bc78160`.

Direct emulator extraction did not register the package with avdmanager. The
bundled Android CLI's exact-version `sdk install emulator@37.1.11` registered it,
using the same numbered archive. The CLI bootstrap downloaded its implementation
when help was first requested; independent pinning of that downloaded CLI remains
a historical supply-chain pinning gap. The downloaded implementation was then
captured at `~/.android/bin/android-cli`, version1.0.16261425, SHA-256
`847e24a7d1711561a8739629b59c6e09b5a80dbfd98045d6ce7c661f46ecbc81`.
Future use must verify and invoke this implementation directly, without the
automatic-download wrapper. This does not retroactively verify the first execution.
No CLI dependency is added to repository automation.
The isolated SDK links paul's existing platform tools (ADB34.0.5-debian).

Created `stuffstash-audit-api36` with Pixel6 dimensions,2GiB memory, two cores and
a2GiB data partition. It boots headlessly using SwiftShader on port5580. ADB reports
`sys.boot_completed=1`, Android16 and fingerprint
`google/sdk_gphone64_x86_64/emu64xa:16/BE2A.250530.026.F3/13894323:userdebug/dev-keys`.
The inspected initial screenshot shows the system wallpaper/status/navigation
bars; remote capture is `~/.cache/stuffstash-android-audit/android16-boot.png`.
Approximately10GiB disk remains after removing the verified archive copies while
retaining their checksum manifests and installed contents.

No Android app has been built or exercised. The existing untracked
`apps/mobile/android/` remains untouched. Android UI acceptance remains an explicit
gap, not a pass inferred from emulator boot.

The fixture installer now supports an explicitly marked disposable source archive
outside GitHub Actions. Root identity, marker contents, Git ancestors, external
backup location and symlink-free route paths are checked before mutation. Six
remote regression cases pass, including preserving an external checkout reached
through a symlink; that case failed before the guard correction. Remote mobile
structural checks pass and critic review has no remaining blocker. This prepares
safe synthetic app installation; it is not an Android app build or UI result.

The disposable source archive at db8d2bb8 installed its frozen mobile dependencies
and completed Android-only Expo prebuild on paul. The generated Gradle9.0.0 wrapper
now has its published distribution SHA-256 configured before execution:
`8fad3d78296ca518113f3d29016617c7f9367dc005f932bd9d93bf45ba46072b`.
Source: [Gradle distribution checksum](https://services.gradle.org/distributions/gradle-9.0.0-bin.zip.sha256).
React Native0.83.6's catalog selects API36, build-tools36.0.0, NDK27.1.12297006
and Android Gradle plugin8.12.0. The initial build failed before app compilation:
Foojay0.5.0 attempted to provision Java17 and referenced `IBM_SEMERU`, removed by
Gradle9. The captured stack trace matches the
[upstream React Native issue](https://github.com/react/react-native/issues/55781).
Installed Temurin17.0.16+8 from its numbered Linux x64 HotSpot archive after checking
published SHA-256 `166774efcf0f722f2ee18eba0039de2d685b350ee14d7b69e6f83437dafd2af1`.
NDK27.1.12297006 installation also completed. The x86_64 build is now retrying with
explicit Java17 and automatic Java downloads disabled, with caches outside the
archive. Its generated release variant uses the debug signing key;
it is not a distributable release. Prebuild also warns that automatic Android
appearance needs expo-system-ui; this needs runtime investigation before a finding
is accepted or a dependency change proposed.
