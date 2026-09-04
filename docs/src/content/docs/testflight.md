---
title: Release To TestFlight
description: Configure Apple cloud signing, publish a tagged release, and check the first TestFlight build.
---

Stuff Stash uploads an iOS build as part of its trusted release workflow on
`main`. That workflow creates the GitHub Release, checks out its exact tag,
creates a native Xcode archive, verifies it, and uploads it to App Store
Connect. It does not publish the app to the public App Store.

The permanent iOS bundle ID is `org.stuffstash.mobile`.

## One-Time Apple Setup

Accept the current Apple Developer agreements before creating the app or
uploading a build. Then:

1. [Register an explicit App ID](https://developer.apple.com/help/account/identifiers/register-an-app-id)
   for `org.stuffstash.mobile` in Certificates, Identifiers & Profiles.
2. [Create the iOS app record](https://developer.apple.com/help/app-store-connect/create-an-app-record/add-a-new-app/)
   in App Store Connect with the same bundle ID. The SKU is an internal value;
   choose a stable unique value such as `stuffstash-ios`.
3. Leave **Associated Domains** disabled for the general TestFlight build. It
   has no fixed invitation host. Enable that capability only for a future
   deployment-specific build that claims links from one verified HTTPS host.

### Create The Automation Key

The workflow uses an App Store Connect team API key for both Xcode cloud
signing and upload. This key has broad account access, so create a dedicated
key and store it only in GitHub and your password manager.

1. In App Store Connect, open **Users and Access → Integrations → App Store
   Connect API**. The Account Holder must
   [request API access](https://developer.apple.com/help/app-store-connect/get-started/app-store-connect-api/)
   first if it is not already enabled.
2. Open **Team Keys**, then choose **Generate API Key**.
3. Give the key a clear name such as `Stuff Stash GitHub Releases` and select
   **Admin** access. The native workflow needs permission to upload builds and
   use Apple's cloud-managed distribution certificates. A lower-access or
   individual key cannot satisfy this signing path.
4. Record the **Key ID** and **Issuer ID**.
5. Download the `AuthKey_<KEY_ID>.p8` private key. Apple allows this download
   only once. Keep the original in a secure credential store; never commit it.

Xcode creates or reuses an Apple cloud-managed distribution certificate during
the first signed archive. See Apple's notes on
[cloud-managed certificates](https://developer.apple.com/help/account/certificates/cloud-managed-certificates/).

If this key is exposed, revoke it in App Store Connect immediately and replace
all three GitHub secrets.

## Add The GitHub Secrets

Open **Repository settings → Secrets and variables → Actions** and add these as
repository secrets:

| Secret | Value |
| --- | --- |
| `APP_STORE_CONNECT_API_KEY_BASE64` | Base64-encoded contents of the downloaded `.p8` file |
| `APP_STORE_CONNECT_KEY_ID` | Key ID shown beside the team API key |
| `APP_STORE_CONNECT_ISSUER_ID` | Issuer ID shown on the App Store Connect API page |

GitHub documents the same UI under
[Creating secrets for a repository](https://docs.github.com/actions/security-for-github-actions/security-guides/using-secrets-in-github-actions#creating-secrets-for-a-repository).

The trusted Release workflow passes these three secrets explicitly to the
reusable TestFlight workflow. The TestFlight workflow does not inherit the
repository's other secrets.

Encode the `.p8` without printing it or creating another credential file. If
GitHub CLI is authenticated for this repository, run:

```sh
base64 < /secure/path/AuthKey_KEY_ID.p8 \
  | tr -d '\n' \
  | gh secret set APP_STORE_CONNECT_API_KEY_BASE64
```

Otherwise, pipe the encoded value to the clipboard on macOS, paste it into the
GitHub secret form, and clear the clipboard afterward:

```sh
base64 < /secure/path/AuthKey_KEY_ID.p8 | tr -d '\n' | pbcopy
# Paste and save the GitHub secret, then clear the clipboard.
pbcopy < /dev/null
```

Do not paste the key into a shell argument, terminal output, issue, workflow
log, or repository file. The release job decodes it into the runner's temporary
directory, restricts its file permissions, and removes it even if the job
fails.

## Server Setup Happens In The App

The general TestFlight build contains no Stuff Stash API URL, tenant hint, or
invitation origin. On first launch, the user enters the URL of their Stuff
Stash server. The app reads that server's public OIDC configuration and
continues sign-in. Changing servers does not require a new app build.

The release workflow explicitly clears those three values and disables
insecure invitation origins, voice diagnostics, and local upload targets. The
app configuration also rejects a general distribution build if any deployment
seed or developer option is enabled, so unexpected runner configuration stops
the release instead of leaking into the binary.

Invitation acceptance for arbitrary self-hosted servers remains in the
browser. iOS universal-link hosts must be declared when an app is built, so one
general binary cannot claim every self-hosted domain.

## Publish A Build

1. Merge the release-worthy Conventional Commits into `main`.
2. The trusted **Release** GitHub Actions workflow runs its required checks,
   chooses the next version, creates the tag and GitHub Release, then calls the
   reusable TestFlight workflow.
3. Watch the **Release** run and its **Publish iOS release to TestFlight** job.
   After it succeeds, wait for Apple to process the upload in App Store
   Connect.

The TestFlight workflow is callable only from the checked-in Release workflow
on `main`. Manually publishing a separate GitHub Release does not trigger an
iOS upload. The Release workflow may also be started with **Run workflow**, but
it publishes only when its release plan finds qualifying commits.

The tag is the source of the user-facing version: `v0.15.0` produces version
`0.15.0`. The workflow accepts only exact `vMAJOR.MINOR.PATCH` tags. The Apple
build number is the Release workflow's run number and attempt, written as
`run_number.run_attempt`. Retrying a run therefore keeps the marketing version
and receives a new build number.

For conservative Apple compatibility, the workflow accepts a run number of at
most four digits and an attempt of at most two digits. It stops before building
if either value exceeds that bound; the numbering scheme must be reviewed and
updated before releases can continue.

The workflow uses the pinned macOS, Xcode, CocoaPods, Node, and pnpm versions.
Its Release archive excludes the Expo developer launcher, developer menu, and
network inspector. Before upload, the workflow verifies:

- Bundle ID `org.stuffstash.mobile`.
- Marketing version from the release tag.
- Build number from the workflow run number and attempt.
- No Expo developer launcher or developer menu bundles.
- No associated-domain entitlement in the general build.

## First TestFlight Check

Before inviting external testers, install the first processed build on a
physical iPhone and check:

- The app opens directly into Stuff Stash with no developer launcher, menu, or
  network inspector.
- About or diagnostics shows the Git release version.
- Server onboarding accepts a real Stuff Stash instance without a baked-in
  server URL.
- OIDC sign-in returns to the app, then tenant and inventory selection works.
- A photo can be authorized, attached, uploaded, and viewed.
- Microphone permission and one voice interaction work.
- A real invitation can be accepted through the browser.
- Sign out clears the session while preserving the saved server choice.

The upload workflow does not complete Apple's product and review setup. Before
external testing, finish the remaining App Store Connect gates: export
compliance, app privacy answers, beta description, feedback address, review
contact and credentials, **What to Test**, and tester groups. External groups
may require TestFlight App Review.
