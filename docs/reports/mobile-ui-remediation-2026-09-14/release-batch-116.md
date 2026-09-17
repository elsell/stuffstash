# Batch116 — search fix merged; TestFlight publication interrupted

PR159 merged as c4626a0be60774ebb0d086885c2465500334aecb. Required CI35247151149
passed, including iOS dependency resolution. Native35247151136 tested merge
f7b3f995dd2ef0e9f26bba2f48d6dfd759075bbd and accepts all eight installed-patch
search workflows on each device (16/16). Reviewed Place/Settings header and iPad
filter captures are retained in native-full-352471.md. Full outcomes74/92 phone,
84/92 iPad are explicitly not comprehensive acceptance. Thirteen shared-header
checks, TypeScript, structural checks,10 fixture checks and critic passed.

Release35258455389 (run116, attempt1) failed after creating immutable tagv0.24.27
at the merged commit. Images and attestations published; GitHub returned HTTP502
while uploading stuffstash-selfhost.tar.gz.sha256. The attempted GitHub release
390958573 is absent afterward (both release-id and tag lookups404). The tag remains
correct and must not be moved or deleted. TestFlight job was skipped: no116.1
upload occurred, and115.1 remains the last verified build.

API digest:sha256:e2aa8126ab59040eb4c9357e655da01aea32ed761b9be76e16ceed3c8c50becf.
Web digest:sha256:dfab3fe1d38c14729d23db4f450ad11c7e467450071e8cea8a7c5b79fce64f95.

Recovery must complete the same tagged release and invoke the existing signing,
upload and changelog path. Blind retry is unsafe: publish explicitly refuses an
existing tag, and the TestFlight workflow currently has no manual dispatch entry.
Add a reviewed recovery route instead of deleting/recreating the tag. No model
polling loop or native-job restart was used; observer slept120 seconds.

Intended TestFlight notes:
- Fixed search appearing at the bottom instead of the header in browsing, settings, and selection screens.
- Known issue: the custom color picker may occasionally fail to open. Preset colors remain available.

The later eager-observation test improvement is on the follow-up branch and is
not part of the frozen release source. Other audit findings remain separate.
