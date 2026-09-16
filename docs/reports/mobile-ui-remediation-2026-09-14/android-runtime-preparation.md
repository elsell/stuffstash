# Android runtime preparation on paul

September16: ADB is installed, but no devices are connected. No emulator or
sdkmanager was found on PATH or in the inspected common SDK locations. Java21.0.9
is available, and john has read/write ACL access to `/dev/kvm`. The home filesystem
has17GiB free; `/tmp` is a memory-backed filesystem and is not the SDK destination.

Installed only the command-line tools under
`~/.cache/stuffstash-android-audit/sdk/cmdline-tools/15859902` on paul. Google's
numbered archive passed its published SHA-256 check before extraction/execution:
`4e4c464f145a7512b57d088ac6c278c03c9eea610886b35a5e0804e74eedf583`.
Source: [Android tools download](https://developer.android.com/studio).
The package listing succeeds; this tools version warns that sdkmanager is
deprecated in favor of its bundled Android CLI.

Google repository metadata identifies these stable candidates; they are **not yet
installed**. Repository XML is retained in the same remote audit directory.

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

No emulator has booted and no Android app has been built or exercised. The existing
untracked `apps/mobile/android/` remains untouched. Android UI acceptance remains
an explicit gap, not a pass inferred from tool installation.
