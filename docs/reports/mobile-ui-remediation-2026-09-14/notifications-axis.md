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

M56 follow-up replaces inbox Retry/Load more and route-load Retry with the shared
native command adapter. Eleven inbox behavior tests/typecheck/structural pass and
critic found no issue; native adaptation is not certified by callback tests.

M57 follow-up scopes device-setup feedback to the initiating foreground/focus
context. Both delayed enabled and denied outcomes are suppressed after backgrounding;
inactive permission prompts retain feedback. Fifteen settings/setup/session tests,
TypeScript and structural checks pass remotely, with the critic's race resolved.
This is transient attempt feedback, not live OS permission or delivery status.
Native Settings-return and physical permission verification remain open.

## September15 — reminder settings interaction follow-through

Source50642b60. Reviewed R043/R044 and S111–S116 through
NotificationSettingsScreen, both notification routes, ExpirationReminderEditor,
ReminderTimingEditor and TimeZonePicker. These are source decisions and lifecycle
checks; no additional native/physical permission success is claimed.

Defaults and expired reminders are booleans, appropriately expressed as switches.
Type mode is a short exclusive choice (defaults/custom/off) using the native menu
adapter. Keep these in place. Timing combines Off, seven meaningful day presets,
and numeric custom entry; a separate bounded editor is justified by its custom
field and save/cancel semantics. Time zones are a large searchable list, not a
three-value preference; its dedicated native search avoids an unwieldy menu.
These are task-fit judgments under the platform-interaction standard.

Preset timing saves on selection; custom timing validates integer0–3650 and uses
a native Save action. Back cancels the unsaved custom draft. Failed saves retain
the choice and allow retry. Type-policy failures retain a draft with explicit
Retry/Discard rather than silently reverting. Source pending guards lock competing
changes, and screen-level abort ownership suppresses late save completion after
focus departure. Native Back/swipe/keyboard timing remains to verify; component
mounted checks alone do not certify focus behavior.

Counterevidence matters: NotificationSettingsRoute keys ScopedNotifications by
server-state scope, tenant and inventory, and memoizes its session within that
owner. Its editor checks expected inventory scope before showing settings. A
hypothesis that screen-local preferences automatically leak into a new selected
inventory is therefore not established; no redundant screen reset was added.
Invalid links and mismatched inventories expose scrollable recovery text with
native navigation available. Actual denied API boundary tests remain separate.

Push preference and device enablement are separate tasks. Enabling can require
system permission; denial offers system Settings, and transient attempt feedback
is invalidated on background/focus change (M57). This is not a live OS-permission
indicator. The query/session remains authoritative for the saved preference.
Long type labels, timezone strings, custom numeric editing, reduced motion,
VoiceOver/TalkBack order and physical Settings-return still need runtime evidence.

Critic found no overclaims. Late-completion suppression above specifically refers
to successful save/navigation; retained child error presentation after focus
changes is not independently certified by this review.
