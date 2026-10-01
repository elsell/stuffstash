---
title: Store Release Notes
description: Publish shared release highlights to TestFlight and Google Play.
---

TestFlight and Google Play share the same reviewed release highlights.

In a feature/fix/performance commit or squash message, write reviewed highlights:

```text
fix: Improve reconnect behavior

Release notes:
- Reconnect without losing your draft.
- Keep search results up to date.
```

Existing `TestFlight notes:` sections continue to work. Otherwise, notes fall back
to feature/fix/performance subjects. The reader deduplicates highlights and selects
first-parent commits since the preceding stable tag. Play receives those highlights
without Apple's version/build heading or changelog footer. Text over Google's
500-Unicode-character limit is shortened with an ellipsis. Prefer concise reviewed
bullets so both stores show all changes. Preview with
`node scripts/play-store-notes.mjs --preview` after setting the generated
`STUFF_STASH_MOBILE_RELEASE_TAG` environment variable.

Enable the Android Publisher API and grant a service account release-management
access to the intended app in Play Console. In the protected GitHub `play-store`
environment, configure secret `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON` with its JSON key.
Keep keys outside Git. Set variable `PLAY_PACKAGE_NAME` if it differs from the
app’s Android package. The key is exchanged for a short-lived scoped token;
credential material and provider error bodies are never printed.

The app and Android release must already exist on the chosen Play track. Run
**Play Store release notes** from `main`, with a published stable GitHub tag, exact
Android `versionCode`, track and locale. The tagged commit must include the Play notes scripts. The workflow reads that immutable commit's notes. It preserves rollout
status/fraction, other releases and other locales; it does not upload an AAB or
promote a release. A release containing several version codes is rejected to
avoid updating a broader release than requested.

For automatic delivery, call the reusable workflow after your Android upload job
has assigned the build to its track. Pass the upload job's exact version-code
output; never query "latest". For example, inside your main-branch delivery workflow:

```yaml
  play-notes:
    needs: android-delivery
    uses: ./.github/workflows/play-store-notes.yml
    with:
      release_tag: ${{ needs.android-delivery.outputs.release_tag }}
      version_code: ${{ needs.android-delivery.outputs.version_code }}
      track: internal
    secrets: inherit
```

`android-delivery` is your product's upload job, not an included binary-publication
job. Use `production` only when that exact release is already assigned there.
Run notes publication serially with other Play automation and avoid simultaneous
Play Console edits. Google's edit commit can submit changes for review, and Play's
review/managed-publication settings control when users see them. Publication
verifies saved notes through a fresh edit; it cannot guarantee review completion.

A rerun with matching notes makes no commit. Failed or uncertain writes are not
retried automatically; inspect Play state and rerun the same exact target. Edits
are cleaned up where possible and otherwise expire. Tests exercise a controlled
Publisher API; actual Play publication requires your configured account.

References: [Publisher tracks](https://developers.google.com/android-publisher/api-ref/rest/v3/edits.tracks),
[edit lifecycle](https://developers.google.com/android-publisher/edits), and
[release-note limits](https://support.google.com/googleplay/android-developer/answer/9859348).

TestFlight notes publish automatically after a successful upload. To retry notes
without rebuilding, dispatch **TestFlight changelog** from `main` with the same
release tag and exact uploaded build number.
