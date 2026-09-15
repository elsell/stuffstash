# Notification-axis review checkpoint

Source373d25d5 plus the M55 candidate. All140 inventoried surfaces are mapped in
`notification-surface-ownership.csv`. This is ownership/applicability analysis,
not140 passing tests. Shared delivery can interrupt every task: root response
handling is reviewed atR006/S130; each task still needs its own lifecycle/draft
acceptance. No unrelated cell has been silently marked passed or N/A.

## Reviewed flow

- `AppServicesContext` installs registration reconciliation and performs push
  cleanup before sign-out/server change. The root installs push navigation after
  services and navigation are ready. Background reconciliation reads permission;
  explicit device enablement owns the permission request.
- `ExpoPushNotificationResponses` handles launch/live precedence, in-flight
  duplicate responses, later deliberate repeat taps and disposal. `OpenPushNotification`
  validates server/principal/routing hints, resolves the current authorized item
  through inbox queries and selects the inventory before returning its identity.
  Payload asset IDs do not directly choose the destination.
- Home's entry is scoped by service/tenant/inventory. The accessible bell label
  distinguishes loading, unavailable count and the unread total; its command
  remains available when count loading fails.
- The inbox offers All/Unread, read/unread, mark-all, paging, explicit error retry,
  and separate location opening. Scoped read/count reconciliation is preserved.
  M55 found and repairs navigation after the originating focus session ends.
- Reminder settings distinguish inventory preference from device permission,
  preserve save failures, offer device Settings after denial, and keep in-app
  reminders separate from push delivery. The device-status message still needs
  live foreground/revocation validation.

## Evidence and remaining findings

Seventy-two remote tests across17 files passed for notification application/adapters,
inbox/settings/bell, followed by TypeScript and structural checks. Three new M55
cases failed first; the final11 inbox tests additionally prove a fresh tap works
after an older completion following blur/refocus. These checks are source/controlled
boundary evidence; they are not APNs delivery or native navigation acceptance.

M55: a delayed inbox open previously navigated after leaving or returning. The
candidate retains completed read state but limits navigation to its original
focus session. Critic found no blocker; its requested fresh-open recovery test passes.

Follow-up inspection is still required for notification Settings after OS permission
changes, native inbox navigation interruption, system foreground presentation,
physical cold/warm pushes, badges, assistive technology and Android. Existing user
push confirmation is historical evidence, not verification of every current path.

Additional source concern: inbox Retry/Load more and route-load Retry still use
bespoke Pressable command styling despite the shared native command adapter. Keep
this task/pattern review open; native adaptation is not certified by callback tests.
