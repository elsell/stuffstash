# Creation workflow coverage after TestFlight 140.1

Reviewed product revision77d129c5, shipped as0.25.0 (140.1). These are nested forms,
not new routes. Route enumeration alone therefore missed them. This review adds
S145–S147 without turning72 surface/axis combinations into independent tests.
It also corrects two stale secondary source references for the inventory-list and
location-list routes after their move into the tab stacks. All current routes are
represented and every listed source exists. This establishes no new user-facing
defect and authorizes no new remediation.

| Surface | Task and native pattern | Commit, cancel and scope ownership |
| --- | --- | --- |
| S145: Edit → Tags → New tag | A separate named form for name and optional color, as requested by the user. It replaces the selection content within the existing visit; Cancel returns to selection. | Add tag updates the selection's pending tags. Done transfers them to the asset draft; only asset Save persists them. The selection owner rejects obsolete scope, disabled and retired completions. Existing and pending duplicate names use the shared resolution command. |
| S146: switcher → New household | A named creation form within the switcher, with a native field and explicit Create/Cancel actions. | Success returns to the created household's inventory list. It does not change active inventory. Pending submission locks duplicate creation and editing; failure retains the name. Blur/scope changes retire presentation ownership. |
| S147: switcher → New inventory | The same form and footer, with the target household identified as context. | The entry requires household creation authority; the command rechecks that authority before the API request. Success shows the new inventory, and a separate selection changes active inventory. Pending, retry and retirement use the same owner as household creation. |

## Evidence and limits

Source inspection: `NewAssetTagScreen`, `AssetTagSelectionScreen`,
`AssetTagSelectionTask`, `WorkspaceCreationForm`, `TenantSwitcherSheetScreen`,
`CreateWorkspace` and `ApiWorkspaceCreation`.

Both forms reuse `NativeFilterSheet`, `DraftTextField`, Settings sections and
native footer commands. They expose explicit field labels, disabled state,
validation/progress or error text, and cancellation. The tag color picker is the
existing adapter. Workspace creation invalidates directory data without changing
the connection selection. These are source observations, not guarantees about
screen-reader traversal, every gesture or system interruption.

Native36167444034 passed the Edit-tag and switcher-creation journeys on iPhone17
and iPad mini. Reviewed normal-text captures cover the separate New tag form,
return to staged selection, retained selection on reopen, household keyboard
entry with reachable actions, and the created inventory in the switcher.
The inventory form shares the household form implementation; the passing creation
journey covers its submission, not every keyboard/window combination. Evidence
and shipped revision are in [release140](evidence/release-140.txt).

The API boundary tests cover legitimate creation and rejection of anonymous,
malformed, cross-household, forged-header and inventory-owner escalation attempts.
Native fixtures use controlled data; they do not certify a physical device's
production API connection. Source tests separately cover failed creation/retry,
duplicate rejection and retired callbacks.

Remaining verification: enlarged text, RTL/long translated content, VoiceOver/
TalkBack traversal, reduced-motion transitions, Android runtime and arbitrary
window sizes. Palette/native adapter use is source-reviewed; the named light
captures do not establish dark/high-contrast acceptance. Search, notifications
and media do not belong to these nested forms; their owning surfaces retain those
obligations. New-tag color editing retains the existing picker's separate coverage.

No further diagnostic run is justified by these inventory additions. The next
product change should follow a user-confirmed connected-workflow issue, with
normal-text structure and hierarchy taking priority over adaptation details.
