# Client Settings Management Spec

## Purpose

Stuff Stash needs one predictable settings entry point on web and mobile, with production-grade management for tenant and inventory customization.

This spec defines the shared user mental model, navigation, interaction states, accessibility behavior, and frontend boundaries for settings, custom fields, custom asset types, and tags. Web and mobile must offer the same supported operations and explain scope and permissions the same way while using platform-appropriate controls.

## Scope

This spec covers:

- One account-oriented Settings entry point on web and mobile.
- Tenant and inventory settings drill-ins.
- Tenant- and inventory-scoped custom field management.
- Tenant- and inventory-scoped custom asset type management.
- Inventory-scoped tag management.
- Active and archived collection views, creation, compatible editing, archive, restore, and permanent deletion where the existing domain supports them.
- Loading, empty, saved, validation, failure, denied, and stale-data states.
- Web and mobile functional parity, accessibility, adapter boundaries, observability, and verification.

This spec does not change domain scope, permissions, lifecycle endpoints, or compatibility rules. In particular:

- Tenant custom fields and custom asset types flow down additively to child inventories.
- An inventory may add its own definitions but cannot override, shadow, rename, archive, restore, or delete an inherited tenant definition.
- Tags remain inventory-scoped. Tenant tags, tag inheritance, tag restore, tag merge, and tag hard delete are out of scope.
- Custom field type and key changes, option removal or reordering, applicability narrowing, target removal, custom asset type key changes, and scope changes remain out of scope.

The domain behavior remains owned by `specs/assets/flexible-asset-fields.spec.md`, `specs/assets/custom-asset-types.spec.md`, `specs/assets/asset-tags.spec.md`, and `specs/platform/resource-lifecycle.spec.md`.

## Product Mental Model

Settings is one destination reached from the signed-in account affordance. It is not a collection of unrelated tools placed beside normal inventory work.

Inside Settings, users choose the level they intend to configure:

- `Account and app` for personal preferences, connection, diagnostics, and about information already supported by the client.
- The current tenant, shown by its user-visible name, for settings shared by that tenant.
- The current inventory, shown by its user-visible name, for settings belonging only to that inventory.

Product copy must use the actual tenant and inventory names. It may use `Tenant settings` and `Inventory settings` as secondary category labels, but ordinary task copy must not require users to understand API scope, inheritance implementation, or authorization internals.

The settings overview must make the containment relationship visible: the current inventory belongs to the current tenant. Switching tenant or inventory must update the available settings destinations and permissions through the same client context boundary used elsewhere in the app.

## Entry Points And Navigation

### Web

- The account row remains at the bottom of the desktop sidebar.
- Its account menu must identify the signed-in account and provide `Settings` and `Sign out`.
- `Settings` in the account menu is the canonical entry point. A separate top-level `Tools` settings destination must not compete with it.
- At narrow web widths, the account control must expose the same Settings destination through the shared mobile account sheet.
- Settings and every drill-in must use durable, directly addressable routes. Back, forward, refresh, and copied URLs must preserve the selected settings level, resource collection, lifecycle view, and selected record.
- Existing settings deep links may redirect to the new canonical route, but two independently navigable settings systems must not remain.

Canonical web route shape:

- `/settings` for the settings overview.
- `/settings/account/{section}` for supported account and app settings.
- `/settings/tenants/{tenantId}` for tenant settings.
- `/settings/tenants/{tenantId}/fields` for tenant custom fields.
- `/settings/tenants/{tenantId}/asset-types` for tenant custom asset types.
- `/settings/tenants/{tenantId}/inventories/{inventoryId}` for inventory settings.
- `/settings/tenants/{tenantId}/inventories/{inventoryId}/fields` for the effective custom field view.
- `/settings/tenants/{tenantId}/inventories/{inventoryId}/asset-types` for the effective custom asset type view.
- `/settings/tenants/{tenantId}/inventories/{inventoryId}/tags` for tag management.

Lifecycle view and selected-record subroutes or query state must be canonical and normalized by focused route helpers. State-changing confirmation routes must remain durable until the operation succeeds or is cancelled.

### Mobile

- Settings remains a native stack destination opened from the trailing account affordance on Home. It must not consume a bottom-tab slot.
- The account affordance must retain the explicit `Open account and settings` accessibility label and a minimum 44-by-44-point target.
- The first Settings screen must use a native grouped-list hierarchy. It must show account/app rows, a tenant settings row labeled with the current tenant name, and an inventory settings row labeled with the current inventory name.
- Tenant settings and inventory settings must push native stack screens. Custom fields, asset types, tags, list rows, and edit forms must continue to push focused screens or use platform-native sheets where the task is short and cancellation is clear.
- Mobile must use native navigation bars, back behavior, grouped lists, menus, sheets, text inputs, switches, pickers, destructive confirmation actions, keyboard avoidance, safe areas, Dynamic Type, and platform accessibility semantics where React Native and Expo support them.
- Full-screen forms are preferred over deeply nested or long bottom sheets. A sheet must not contain another sheet.
- Leaving a dirty form by back gesture, back button, or dismiss action must prompt to discard changes. Clean forms dismiss without confirmation.
- Dirty-form protection in a native stack must participate in the navigator's prevent-remove protocol before the native screen is removed. `Keep editing` leaves the populated editor mounted and dispatches no pending navigation action. `Discard` dispatches the intercepted action exactly once. This applies equally to header back, interactive back gesture, and Android hardware back.
- A dirty native-stack editor must disable the interactive pop gesture before it can begin a visual transition. Its visible header back or close action must invoke the same prevent-remove path, and normal gesture behavior must be restored as soon as the editor is clean or unmounts. A directly opened editor must always show a labeled, visible exit affordance and return deterministically to its owning collection.

### Mobile Settings Layout Contract

- Mobile settings screens use one shared 16-point horizontal content inset for collection search, Add, lifecycle controls, notices, grouped rows, form actions, lifecycle actions, and empty or error recovery content. A primary action and the lifecycle group beneath it must align to the same content column and width.
- Related groups use the shared settings section rhythm rather than screen-specific edge spacing. Controls must retain at least a 44-by-44-point target, allow labels and values to wrap under Dynamic Type, avoid clipping at narrow widths, and keep the final action clear of the bottom safe area and keyboard.
- A collection's search, Add action, lifecycle selector, progress notice, and rows must read as one aligned surface. None of that collection chrome may touch the viewport edge while its rows are inset.
- The collapsed mobile tag color control must keep `Custom…` visible without horizontal scrolling. Preset colors may wrap or use a bounded grid, but arbitrary-color discovery cannot depend on a horizontally clipped trailing control.
- The custom-color modal uses the same horizontal inset, bottom safe-area clearance, fixed non-scrolling picker area, and aligned full-width actions. Its Cancel, Done, and Clear controls remain at least 44 points and usable with the keyboard visible.
- Read-only, inherited, and archived mobile detail uses static labeled values rather than disabled text inputs, color pickers, or other controls that imply mutation.
- Mobile create forms identify required values neutrally on first presentation. Field-level errors become assertive only after the user has interacted with the affected control or attempted submission; a blank untouched form must not open by announcing errors.
- Custom-field type and applicability use compact value/disclosure rows that open a focused single-selection surface. The editor must not render a viewport-tall grid of radio cards for these single-value choices.
- The custom-color surface may scroll its overall content at large Dynamic Type or while the Android keyboard is visible, but the spectrum and hue controls retain gesture ownership and do not scroll during color gestures.

## Settings Information Architecture

The settings overview must be list-first and compact. It must not lead with creation forms, dashboard metrics, or instructional cards.

Tenant settings initially includes:

- `Custom fields`.
- `Asset types`.
- Other supported tenant administration, such as Voice Setup, may remain a peer destination without being mixed into customization screens.

Inventory settings initially includes:

- `Sharing`.
- `Tags`.
- `Custom fields`.
- `Asset types`.
- Other supported inventory administration and activity destinations may remain peers.

Each row must have a familiar icon, visible label, optional concise state summary, and disclosure affordance. Icons supplement labels and must not be the only way to distinguish tenant from inventory.

Unavailable sections must not disappear merely because the caller lacks mutation permission when read access still supports a useful view. The destination must show readable content with clear read-only treatment. A destination the caller cannot safely read must show a calm denied state without leaking hidden names or identifiers.

## Shared Collection Pattern

Custom fields, custom asset types, and tags must use the same high-level collection mental model on web and mobile:

1. A scannable collection is the primary screen.
2. An `Add` action opens a focused creation flow when allowed.
3. Selecting a row opens detail/edit for that record.
4. `Active` and `Archived` are distinct views when the domain exposes archived records.
5. Destructive or lifecycle actions are secondary to ordinary editing.

The shared mental model does not require shared web/mobile UI code. Clients should share domain vocabulary, behavior scenarios, and semantic design tokens where useful while preserving native mobile behavior and shadcn-svelte web composition.

Collection requirements:

- Active collections sort by localized, case-insensitive display name with stable ID as a tiebreaker unless a domain endpoint specifies another authoritative order.
- Clients must follow cursor pagination until the first bounded screen is available and expose `Load more` or platform-appropriate incremental loading for remaining pages. They must never describe a partially loaded collection as complete.
- Pull to refresh is required on mobile collection screens. Web requires an explicit retry and may expose refresh when stale data remains visible.
- Replacement loading must preserve the last successful rows. Initial, replacement, and pagination failures need distinct safe recovery states.
- Mobile lifecycle switching must use the shared settings segmented-control component. It must render the platform-native segmented control supplied by the pinned Expo UI runtime on iOS and Android when the required native views are available, with an accessible project fallback for incompatible binaries. The selected segment's shape and state semantics are sufficient; it must not add a redundant checkmark. `Active` and `Archived` data must commit atomically with the selected segment so rows are never shown under the wrong lifecycle.
- Compact replacement progress must use the shared inline settings-loading component with the activity indicator and label side by side. It must not stack the spinner above its text or replace the last successful collection during a lifecycle transition.
- Search is required when a collection contains more than twelve records or when more pages are available. Search filters already loaded rows immediately and must use an API search contract before claiming to search unloaded pages. Until such a contract exists, the client must finish loading the collection before treating local search as complete, with visible loading/progress and bounded pagination-loop protection.
- Empty states must distinguish `No active ...`, `No archived ...`, no search matches, permission denial, and load failure. Add is offered only in the active empty state and only when permitted.
- A stale deep link to an archived or unavailable record must produce a safe not-found or denied state and a route back to the collection.
- Each row must have one primary selection target. Secondary actions belong in a native menu or accessible overflow menu, not a row of competing icon buttons.
- Lifecycle state, inherited/local ownership, and availability must not rely on color alone.

## Custom Fields

### Effective Inventory Presentation

An inventory custom-fields screen must show two labeled groups:

- `From {tenant name}` for inherited tenant-scoped definitions.
- `Only in {inventory name}` for inventory-scoped definitions.

Inherited definitions are available to assets in the inventory but are not owned by the inventory. They must have an explicit `Inherited` label and read-only presentation in the inventory screen. They must never show inventory edit, archive, restore, or delete actions.

If the caller has `tenant.configure`, an inherited record may offer `Manage in {tenant name}` and navigate to the tenant-scoped record. Otherwise it remains readable without a disabled mutation control. The client must not infer tenant permission from inventory permission.

Tenant and inventory lists must expose concise metadata: display name, field type, applicability, and lifecycle state when archived. Stable key is secondary technical detail, not the primary row title.

### Create And Edit

- Add Field is a focused form, not a permanently expanded panel above the list.
- Scope comes from the settings level and is not an editable form choice.
- Create supports display name, stable key, type, enum options when applicable, applicability, and eligible custom asset type targets.
- The client may derive a suggested key from the display name. While the user has not explicitly edited the stable-key control, the suggestion must continue to track the complete display name as it is typed rather than freezing after the first character. Explicit stable-key editing stops automatic replacement so the user's reviewed value is preserved. The key must remain reviewable before creation and immutable afterward.
- Field and asset-type clients must use the same stable-key validator as their application managers. If a generated or manually entered key is invalid while technical details are collapsed, the form must reveal and focus that control, explain the accepted format, and keep Save unavailable.
- Edit supports only compatibility-preserving operations already defined by the domain: display-name change, adding enum options, adding active eligible custom asset type targets, and expanding targeted applicability to all assets.
- Immutable values must render as read-only details. Unsupported narrowing or removal controls must not be rendered as disabled editable controls.
- Enum option and custom asset type target adders must prevent duplicate selections and explain why archived or wrong-scope targets are unavailable without exposing unauthorized records.

### Lifecycle

- Active definitions may be archived after an explicit confirmation naming the field and explaining that existing stored values remain but the field is hidden from normal editing and validation.
- Archived definitions have read-only detail with `Restore` and `Delete permanently` when permitted.
- Restore must surface target-revalidation failures safely and leave the definition archived.
- Permanent delete must use a second, clearly destructive confirmation that explains irreversibility and that deletion can be blocked while active assets store a non-empty value.
- A blocked delete must preserve the archived record and explain the next action without exposing asset data the caller cannot view.

## Custom Asset Types

### Effective Inventory Presentation

Inventory asset-type screens use the same `From {tenant name}` and `Only in {inventory name}` grouping, inherited read-only behavior, and optional `Manage in {tenant name}` navigation as custom fields.

Rows must show display name and concise description when present. Stable key and lifecycle state are secondary details.

### Create And Edit

- Add Asset Type is a focused form with display name, stable key, and optional description.
- Scope comes from the current settings level.
- A suggested key may be derived before creation. It must continue to track the complete display name until the user explicitly edits the stable-key control, after which the user's reviewed value is preserved. It becomes immutable after creation.
- Edit supports display name and description only.
- ID, tenant, inventory, scope, and key remain read-only details and must not look editable.

### Lifecycle

- Active asset types may be archived after confirmation that existing asset and field references remain while new assignment and targeting become unavailable.
- Archived asset types are read-only and may expose `Restore` and `Delete permanently` when permitted.
- Permanent-delete confirmation must explain irreversibility and that deletion is blocked while active assets or custom field targets reference the type.
- Blocked lifecycle operations preserve the prior state and provide safe recovery copy.

## Tags

- Tags appear only under the current inventory.
- The tag manager lists active tags alphabetically with display name, accessible color treatment, and optional assignment count only if an authoritative count exists. Clients must not derive or imply a complete count from a partial asset list.
- Every tag-management row on web and mobile must reserve the same leading color-indicator slot. A colored tag shows its color as a filled circle; a tag without a color shows an empty outlined circle with sufficient contrast. The reserved slot must preserve row alignment regardless of whether a color exists.
- The color indicator supplements the tag name and must expose a non-color accessible name or value such as `Blue color` or `No color`. Color must never be the only way to identify, select, or distinguish a tag.
- Tag-management collection rows must remain compact and show the tag name as their only visible text. They must not repeat the already-visible color indicator as a `No color`, color-name, or hex-value subtitle. Assistive technology must still receive the color state through the row's accessible name or value.
- Add Tag is a focused form or short native sheet with display name and optional color. Key derivation and normalization stay behind the client application boundary; ordinary UI does not make the stable key a required decision unless resolving a collision requires it.
- Edit supports display name and setting, changing, or clearing the optional color.
- Web tag create and edit must provide the browser-native color picker through an accessible native color input. iOS tag create and edit must use the native SwiftUI color picker through the pinned Expo UI integration. Android, which has no equivalent picker in the pinned Expo SDK, must use an accessible project-owned full-spectrum picker that can choose arbitrary valid `#RRGGBB` colors and includes labeled hex entry as a fallback. A swatches-only mobile control is not sufficient on either platform.
- All platforms may offer labeled quick swatches in addition to the full picker and must provide an explicit accessible `Clear color` action. The selected value must also be communicated with text or selected-state semantics, not only by the displayed color. Hex entry may be offered as an advanced web fallback but must not replace the browser-native input.
- Mobile color selection stays compact until requested: the form shows `No color`, preset swatches, and a clearly labeled `Custom…` control that preserves an arbitrary current color as a visible selected indicator. It must not render the full-spectrum surface inline in the scrolling form.
- `Custom…` opens a dedicated native modal or sheet. A supported iOS binary presents the SwiftUI picker in that surface. Android and iOS binaries without the required ExpoUI native views present the project full-spectrum picker in a fixed, non-scrolling interaction area with labeled hex entry, `Clear color`, cancellation, and `Done`. The spectrum and hue surfaces must retain responder ownership for the gesture so the editor behind the modal cannot scroll instead. On a keyboard-shortened viewport or at large Dynamic Type, the fixed picker must use a measured compact presentation while preserving direct hue, saturation, and brightness access; the supplementary region remains scrollable and cancellation and `Done` remain pinned above the keyboard.
- `@expo/ui 55.0.17` is the pinned iOS SwiftUI color-picker bridge. Its platform behavior, accessibility support, compatibility with the pinned Expo SDK, and supply-chain posture must be verified under `specs/platform/tooling-versions.spec.md`; it must not be presented as providing a native Android color picker.
- Archive is the only tag lifecycle mutation in this slice. Confirmation must explain that the tag will no longer be available for new assignment or normal filtering while audit and existing history remain.
- Tag management is active-only in this slice. After successful archive, the tag leaves the active collection. Clients must not fake an Archived view from local state or history and must not expose Restore or Delete permanently because the tag domain does not define those operations.
- Inline tag creation and asset tag assignment remain available in asset create/edit. They must use the same frontend tag domain model, validation, color treatment, and adapter behavior as the manager.

## Permissions And Denied States

Clients must use effective permissions returned by the API for presentation and let server authorization remain authoritative.

- Tenant custom field and custom asset type list/create/update/archive/restore/delete require `tenant.configure` under the current API contract.
- Effective inventory custom field and custom asset type lists require `inventory.view`.
- Inventory-owned custom field and custom asset type mutations require `inventory.configure`.
- Tag list requires `inventory.view`.
- Tag create/update/archive requires `inventory.edit_asset`.

Read-only users must be able to inspect effective inventory definitions and active tags when list permission permits. Mutation controls must be omitted rather than presented as unexplained disabled controls. A direct mutation deep link without permission must render a denied state, never a briefly enabled form or raw server error.

Permission changes during an open screen or save must fail closed, preserve unsaved user input when safe, refresh effective permissions, and explain that the change was not saved.

## Forms, Feedback, And Safety

- Forms must have persistent visible labels, field-level validation, a concise error summary for submission failures, and programmatic label/error associations.
- The primary save action is disabled only when the form is invalid, unchanged, or saving, with a visibly unavailable treatment.
- Save is single-flight. Repeated taps or clicks must not issue duplicate mutations.
- Successful create returns to the collection and makes the saved record visible. Successful edit remains on detail or returns according to the platform convention, with an explicit saved notice.
- Saved notices must be announced without stealing focus. Web may use the shared toast/notice system; mobile must use the shared native-feeling notice surface.
- Validation failures preserve entered values and move focus or accessibility focus to the error summary or first invalid field.
- Safe API failures say what was not saved and offer retry. Raw transport messages, stack text, request IDs, provider details, and authorization internals must not enter product copy.
- Archive, restore, and delete actions must be single-flight and update the collection only after server success.
- Cancellation never mutates server state.

## Accessibility And Responsive Behavior

- WCAG 2.2 is the web baseline; Apple Human Interface Guidelines and platform conventions guide native mobile behavior where they are at least as strong.
- Web keyboard order must follow visual order. Every row, menu, dialog, sheet, picker, lifecycle view, and form control must be operable without a pointer.
- Dialogs and sheets must trap focus while open, make the background inert, announce their title, and restore focus to the invoking control on close.
- Mobile must support VoiceOver/TalkBack semantics, Dynamic Type, increased contrast, reduced motion, and screen-reader announcements for saved, failed, denied, and lifecycle changes.
- Meaningful mobile controls must provide at least 44-by-44 points. Web touch layouts must provide at least 44-by-44 CSS pixels.
- Selected lifecycle views, inherited state, chosen targets, selected colors, validation state, and destructive actions must not depend on color alone.
- Long tenant, inventory, field, type, tag, enum-option, and description text must wrap without clipping at narrow widths and large text sizes.
- Web must use responsive lists and focused forms rather than compressing desktop tables onto narrow screens.
- Destructive actions belong below ordinary edit actions and use native destructive styling or the shared destructive web primitive.

## Frontend Boundaries And Shared Components

Both clients must preserve hexagonal boundaries:

- UI collects intent and renders frontend domain models.
- Application commands and queries coordinate list pagination, permission-aware presentation, form validation, save state, lifecycle actions, and navigation destinations.
- API adapters map generated DTOs and safe errors into client-owned models.
- Generated API types must not become form or component state models.
- Authentication, selected tenant/inventory context, runtime configuration, and observability remain behind their existing ports/helpers.

Shared behavior within each client should use focused components:

- Settings grouped-section and destination row.
- Collection header with lifecycle view and Add action.
- Customization row with scope/ownership and lifecycle metadata.
- Empty, loading, denied, recoverable error, and pagination footer states.
- Labeled form field and validation summary.
- Choice grid/picker for field type, applicability, enum options, and custom asset type targets.
- Tag chip and color picker.
- Destructive confirmation and saved notice.

Web generic controls must use the local shadcn-svelte primitives. Mobile should use platform-native components where they express the intended behavior. Cross-platform consistency is defined by information architecture, vocabulary, available actions, state behavior, and semantic tokens, not identical pixels or a forced shared UI package.

Files must remain separated by route, screen/component, frontend domain type, application command/query, port, adapter, mapper, validation, observability, and test responsibility before any becomes a catch-all.

## Observability

Client observability must use an injected helper or port and safe domain-oriented events. Raw `console.log`, `print`, and ad hoc diagnostics are prohibited.

At minimum, clients should emit safe events for:

- Settings opened and settings level selected.
- Custom field, custom asset type, or tag collection load failed.
- Create, update, archive, restore, or delete requested and succeeded or failed.
- Permission denial encountered.

Events may include safe scope kind, resource kind, lifecycle action, client platform, and typed failure category. They must not include credentials, tokens, custom field values, raw descriptions, tag names, emails, raw server messages, or hidden resource identifiers.

## Functional Parity

Web and mobile are complete only when both support every operation available through the generated API and allowed by this spec for the same effective principal:

- Discover Settings through the account affordance.
- Navigate account/app, current tenant, and current inventory settings.
- List active tenant and effective inventory custom fields and asset types.
- Distinguish inherited and inventory-owned definitions.
- Create and compatibly edit tenant- or inventory-owned fields and types.
- Archive, list archived, restore, and guarded hard delete fields and types.
- List, create, edit, and archive inventory tags, without inventing tag restore.
- Render equivalent loading, empty, pagination, search, saved, validation, failure, denied, and stale-data states.

Platform-specific navigation and controls are expected. A visually similar screen that lacks an operation, permission state, lifecycle state, or reliable recovery path is not parity.

## Verification

Implementation must follow test-driven development and verify behavior through client application boundaries with fakes rather than mocks.

Required web and mobile tests:

- Settings entry-point and hierarchy navigation.
- Tenant/inventory context switching and route/deep-link correctness.
- Effective inherited-versus-local grouping.
- Viewer, editor, inventory configure, and tenant configure presentations.
- Direct denied deep links and permission changes during save.
- Cursor pagination, incomplete-list language, search over paginated collections, replacement loading, stale response rejection, retry, and empty states.
- Create and every allowed edit transition for fields, types, and tags.
- Tag rows with and without colors preserve the same leading-slot alignment, expose non-color accessibility semantics, and remain distinguishable in increased-contrast and dark appearances.
- Tag create/edit exposes the native web color input, native SwiftUI picker on iOS, and accessible full-spectrum project picker plus hex fallback on Android; each applies arbitrary picker values, supports any quick swatches without making them the only picker, and clears an existing color through the explicit action.
- Immutable property presentation and rejection of unsupported evolution.
- Archive confirmation and success/failure behavior for all three resource kinds.
- Field/type archived listing, restore, guarded permanent delete, and blocked lifecycle operations.
- Absence of tag restore/delete controls.
- Dirty-form dismissal, duplicate-save prevention, validation focus, saved announcements, and safe error mapping.
- Keyboard/focus behavior on web and VoiceOver/TalkBack labels, Dynamic Type, touch targets, contrast, and reduced motion on mobile.

Web must be verified visually at desktop, narrow desktop/tablet, and phone/reflow widths in light and dark appearance, including long names, empty lists, dense lists, mixed colored and uncolored tag rows, the open native color picker where screenshot tooling supports it, inherited/local groups, denied state, validation errors, open menus/dialogs, and archived lifecycle confirmations. Automated browser checks must cover critical navigation and mutations against a controlled fake or seeded environment and must verify native color-input semantics independently of screenshot support.

Mobile must be verified on iOS and Android where supported, in light and dark appearance, at default and large accessibility text sizes. Screenshot review must cover the settings overview, tenant and inventory drill-ins, each collection, mixed colored and uncolored tag rows, tag create/edit with the native SwiftUI picker opened on iOS and the full-spectrum project picker opened on Android, inherited read-only detail, denied state, validation error, and destructive confirmation. Manual verification must confirm that each platform picker can choose an arbitrary color, return the chosen value, and clear it without relying on quick swatches; Android verification must also cover labeled hex fallback entry. If the primary agent cannot run a mobile build, the product owner may supply screenshots and interaction results, but automated application tests and a documented manual checklist remain required before completion.

### Mobile Manual Verification Checklist

Record the device or simulator, OS version, appearance, text size, principal role, and pass/fail evidence for every checked item. Repeat the visual and interaction checks on both iOS and Android unless a row names one platform.

- [ ] Open Settings from the Home account affordance and confirm the account, named tenant, and named inventory hierarchy at default and largest accessibility text sizes.
- [ ] Open the tenant and inventory drill-ins and every fields, asset-types, and tags collection in light and dark appearance; confirm safe-area layout, readable long names, 44-point targets, and no clipped content.
- [ ] Verify empty, dense, searching, refreshing-with-stale-rows, retry, incomplete-list, and permission-denied collection states.
- [ ] Verify inherited and inventory-owned groups, inherited read-only detail, and `Manage in {tenant}` availability for tenant configurators only.
- [ ] For viewer, editor, inventory configurator, and tenant configurator principals, confirm readable destinations remain visible and mutation controls appear only when authorized.
- [ ] Revoke access during collection/detail load and during save; confirm explicit denial, no retained protected rows, safe copy, and a populated read-only draft with refresh recovery after a denied save.
- [ ] Create and edit every allowed field, asset-type, and tag transition; verify immutable details, validation focus/announcement, saved announcement, single-flight submission, dirty-form dismissal, and safe API-failure recovery.
- [ ] Archive fields, asset types, and tags; list/restore/guardedly delete archived fields and asset types; confirm blocked operations preserve state and tags never expose restore or permanent delete.
- [ ] Confirm colored and uncolored tag rows retain the same leading slot and announce a human color name or `No color` with VoiceOver and TalkBack.
- [ ] On iOS, open the native SwiftUI picker, choose an arbitrary non-swatch color, confirm the returned hex value is applied, clear it explicitly, and confirm reopening represents no selection rather than a fallback color.
- [ ] Verify an iOS development client rebuilt after adding or updating `@expo/ui` uses the SwiftUI picker. Before that rebuild, or in any binary without the ExpoUI Host and ColorPicker native views, confirm the same screen falls back to the project full-spectrum picker without an unimplemented-component error. Rebuild with `pnpm --dir apps/mobile exec expo run:ios` to restore the native picker.
- [ ] On Android, open the full-spectrum picker, choose an arbitrary non-swatch color by touch and accessibility actions, confirm the returned hex value is applied, clear it explicitly, and repeat selection through the labeled hex fallback.
- [ ] Confirm the collapsed color row contains only `No color`, presets, and `Custom…`; opening `Custom…` presents a modal/sheet, spectrum and hue drags do not move the editor behind it, Cancel preserves the original value, and Done applies one valid value.
- [ ] From a populated dirty editor, exercise header back, interactive back gesture, and Android hardware back. For each, confirm `Keep editing` remains on the populated editor and `Discard` performs the intercepted navigation exactly once.
- [ ] With VoiceOver/TalkBack, increased contrast, and reduced motion enabled, verify row labels, tabs, picker values/actions, notices, denied states, validation errors, and destructive confirmations are announced and operable.

Each implementation pass must run the relevant type checks, unit/application tests, accessibility checks, structural/pre-commit checks, and the code critic agent. Confirmed findings must be fixed; deferred findings must be explicit.

## Open Risks

- Tags have no archived-list, restore, or hard-delete contract. The active-only manager must make archive consequences clear without suggesting recovery that the domain cannot perform.
- Complete client-side search over a paginated settings collection requires loading all pages until a server-side search contract exists; implementations must remain bounded and honest about progress.
- Tenant-scoped customization lists require `tenant.configure`, so non-configuring tenant viewers cannot inspect tenant definitions directly even when they can see inherited definitions through an inventory effective list.
