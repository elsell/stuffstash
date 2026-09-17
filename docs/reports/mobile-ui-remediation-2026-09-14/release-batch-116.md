# Released batch — TestFlight0.24.27 (116.2)

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
upload occurred in that attempt. Recovery completed below.

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


Recovery completed: PR160 merged dde7a3657a80961bdad53442450b35365c153768 after
CI35260165384 passed. GitHub releasev0.24.27 was restored from the exact tagged
source and already-attested image digests; the rebuilt archive checksum verified.
The immutable tag remains c4626a0be60774ebb0d086885c2465500334aecb.

Recovery35261115108 succeeded: validation105336789686, tagged iOS upload
105336855472 and changelog105343149300 all passed. Apple processing and exact
TestFlight0.24.27(116.2) notes readback verified at2026-09-17T19:09:34Z. This is
the released search correction;116.1 was never uploaded. The operational recovery
workflow is main-only and reuses signing, archive checks and exact-build notes.

M207 native search placement is corrected and released for the verified workflows.
The comprehensive audit remains incomplete; normal-text color/text-entry findings
and remaining adaptation/assistive-technology verification stay tracked.
