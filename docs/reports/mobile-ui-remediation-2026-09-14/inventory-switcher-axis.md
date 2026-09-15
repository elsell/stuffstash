# Inventory switcher review

Source064fbaf9; R056 and S080–S083. Reviewed tenant-switcher route, root sheet
configuration, TenantSwitcherSheetScreen, IdentityLabel and SelectInventoryCommand.
This is a source review plus the focused remote regression evidence, not native
visual or assistive-technology acceptance.

## Task, navigation and selection

Switching inventory changes the context used by subsequent tasks. The current
inventory is marked selected; choosing another household first narrows its inventory
list without yet committing a different inventory. A bounded hierarchical chooser
is appropriate for this relationship. Do not flatten household/inventory names
into an unqualified menu: households can share a name and inventories have roles.
The existing M07 test verifies household identity rather than name matching.

The sheet supports medium/full detents, a scrolling body and native Close. The
household list returns to inventories when a household is chosen. Empty households
have an explicit message. Retry stays in the sheet after load failure; selection
failure retains the chooser and gives a safe retry message. This review does not
claim that every custom command or row already uses the best native adapter.
Switch household/Back and load Retry remain custom commands and should be compared
with the existing NativeCommandButton adapter in the next control pass.

## Ownership and recovery

M78 covers focus departure during selection, both orderings of completion/refocus,
and a fresh selection afterward. Pending selection blocks a duplicate request;
Close aborts the initiating request and dismisses. Focus cancellation is not a
rollback promise. M79 ensures repository acceptance still notifies the scoped
selection observer even if the caller cancels before completion. Initially canceled
or rejected requests do not publish selection. Repository/API authorization is
unchanged and requires its separate boundary evidence.

The UI reads cached dashboard data through the inventory-scoped query adapter.
Initial loading and failure are distinct from cached content; selection has separate
busy/error state. Loading copy still says “tenants,” although the product calls
them households. Correct this vocabulary in the control pass. Background-refetch
failure with retained data has no explicit notice here; evaluate its effect on
stale membership before treating it as a proven user-facing failure.

## Layout and accessibility gates

Rows expose selected/disabled state and inventory names; labels and metadata are
separate from the role badge. The body scrolls and uses automatic inset adjustment.
These properties do not prove correct VoiceOver grouping or native target geometry.
The household header uses IdentityLabel's single-line default. Long names and
large type need visual inspection, particularly beside Switch household and the
nonshrinking role badge. Existing full-width row targets are not proof that their
contents fit. Light/dark, contrast, RTL, narrow/iPad windows, reduced transparency,
VoiceOver/TalkBack and interruption by push navigation remain pending.

Critic-reviewed implementation evidence for M78/M79 is in findings.md. No native
switcher run has been inferred from the unrelated sheet-diagnostic scenarios.
