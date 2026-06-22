# Mobile App Tracer Bullet Spec

## Purpose

Stuff Stash needs a minimal mobile app slice so the team can verify Expo Go on a real iPhone before investing in mobile product workflows.

## Scope

This spec covers only the first mobile scaffold:

- A separate Expo React Native app in the monorepo.
- A TypeScript entry point that renders a static confirmation screen.
- Local development commands for Expo Go.

This spec does not define authentication, API calls, navigation, camera behavior, voice interaction, release signing, TestFlight, EAS builds, or production mobile distribution.

## Decisions

- The first mobile app must live under `apps/mobile`.
- The first mobile app must use Expo, React Native, and TypeScript.
- The first mobile app must target Expo SDK 54 for Expo Go physical-device validation during the SDK 56 transition period.
- The first screen must be static and must not call the API.
- The app must not require an Expo account for the first local validation path.
- The app must not add native modules beyond the Expo blank TypeScript template dependencies.

## Requirements

- Mobile dependencies must be pinned exactly in `apps/mobile/package.json`.
- Mobile dependency versions must be recorded in `specs/platform/tooling-versions.spec.md` before use.
- `pnpm --dir apps/mobile start` must start the Expo development server.
- The root package must expose a convenience script for the mobile development server.
- The first screen must make it obvious that the app launched successfully in Expo Go.

## Verification

- `pnpm --dir apps/mobile check` must type-check the mobile scaffold.
- The Expo development server should start with `pnpm --dir apps/mobile start`.
- iPhone verification is manual: install Expo Go, run the mobile dev server, scan the QR code, and confirm the static Stuff Stash screen appears.
