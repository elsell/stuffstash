# Asset command review

## Current S096/S097/S098 follow-through — all24 source axes

Reviewed at46f23136 plus M203. This supersedes the historical confirmation and
blur-coverage observations below. Checkout and Return remain direct native
commands. Restore now also acts directly; asking whether to reverse an archive
added interruption without an additional choice or irreversible consequence.
Apple advises avoiding alerts for routine reversible actions; applying that to
Restore is a project design judgment. Archive's removal from normal work retains
the existing project confirmation, while permanent deletion remains explicitly
irreversible. [Apple alerts](https://developer.apple.com/design/human-interface-guidelines/alerts).

| Axes | Current source evidence and remaining acceptance |
| --- | --- |
| Task, selection | Checkout/Return change availability; Restore returns to active work. These are commands, not value pickers. Permanent delete is segregated from ordinary commands. |
| Navigation, modality | Native iOS UIMenu and Android action-menu adapter expose lifecycle actions. Restore needs no alert. Archive/Delete still name the affected asset and allow Cancel. Successful focused delete returns Back or replaces with Home if no history exists. |
| Layout, adaptation, typography | Availability uses shared native command adapters; menu/alert layout is platform-owned. Normal-size long names, iPad anchoring and source-to-destination focus remain native checks. Enlarged-text work follows normal-size issues. |
| Appearance, imagery | Labels and SF Symbols identify operations; permanent delete has destructive semantics. Shared semantic status palette. Contrast, disabled appearance and icon accessibility traversal remain unverified. |
| Localization | English action/status text includes asset names; long translated text and RTL are unverified. No date/number entry belongs to these direct commands. |
| Targets, gestures, keyboard | Explicit buttons/menu entries avoid gesture-only actions. There is no command-specific text input here; Home's optional return notes are a separate task. Actual header/menu hit areas and hardware keyboard operation remain runtime work. |
| Accessibility, motion | Named controls and working/success text exist. Native menu/alert focus, announcements and reduced-motion behavior need inspection. Prop-level labels are not spoken-output evidence. |
| Content, search | Current asset capability flags choose available actions. Commands have no independent search or paged list; checkout/history navigation leads to separately audited read tasks. |
| Loading | A synchronous resource-owned lock prevents duplicate requests and competing actions; pending UI disables controls. Departed completion cannot unlock a replacement asset's operation. |
| Recovery | Current-visit failure is safe feedback with retry. Mutation success followed by refresh failure is described as successful mutation with refresh guidance, rather than falsely claiming the mutation failed. |
| Editing | No intermediate draft for these direct commands. Restore can be reversed by Archive through ordinary lifecycle policy. These Details actions do not offer Home's return-note/Undo task; no global Undo guarantee is inferred. |
| Privacy | Capability-filtered presentation and injected command ports preserve architectural boundaries. API authorization remains authoritative, and these UI tests do not prove server isolation. Changes to capability while a confirmation is already displayed need separate native/integration review; focus ownership alone does not certify that scenario. |
| Notifications, media | No permission prompt or media acquisition owned here. Delete messaging accounts for attachments/contents consequences. External notification interruption is a route lifecycle scenario, not verified physical behavior. |
| Lifecycle | Source tests cover success/failure after blur/refocus for all five actions, old confirmation retirement for Archive/Delete, duplicate confirmation consumption and resource replacement. Process termination and OS background behavior remain open. |

M203 verification: three cases failed before direct Restore dispatch. The new
command case verifies no added alert, one pending request despite repeated taps,
failure and successful retry. Shared visit tests retain Restore success/failure.
Obsolete Restore-confirmation cases were removed, leaving Archive/Delete coverage.
All115 related tests across5 files pass remotely (`/tmp/restore-direct-green.log`).
The first typecheck found a nullable test-instance argument; normalizing it to
undefined restores static compatibility. Static evidence is recorded separately
in `/tmp/restore-direct-static.log`. Critic found no blocker. Native direct-menu
Restore acceptance remains pending.

Current follow-up: availability and maintenance now use NativeCommandButton;
the custom-Pressable note below describes the earlier checkpoint. The shared
[Details route review](asset-detail-route-axis.md) covers R012/R020 across all24
axes and records M195 photo acquisition failure ownership. Native acceptance
remains open.

Scope: S095 overflow, S096 checkout, S097 return, S098 archive/restore/delete.
Reviewed source at7fcb38cb with the M86 candidate applied. This is source evidence,
not a native interaction pass.

| Axes | Source conclusion and remaining acceptance |
| --- | --- |
| Task, selection | These are commands, not field choices. Overflow groups history navigation, reversible lifecycle actions and permanent deletion. Checkout/return are direct commands. No extra selection screen is needed for the current command inputs. |
| Navigation, modality | iOS overflow is a native header menu; Android uses the native action-menu adapter. History entries navigate. Lifecycle actions use named confirmation alerts, with explicit Cancel and destructive styling only for permanent deletion. Checkout/return do not add confirmation. Native menu dismissal and focus return remain pending. |
| Content, imagery | Action labels and SF Symbols describe their command, with permanent deletion in a separate group. Available actions derive from the current asset's capability fields. Native icon rendering and long asset names need inspection. |
| Editing, recovery, lifecycle | M86: checkout/return/lifecycle lacked the synchronous owner used by photos. Duplicate callbacks submitted twice. The candidate shares the owner, rejects obsolete asset callbacks and suppresses UI effects after unmount/replacement. Domain commands still finish. Tests verify duplicates and old completion isolation. Blur while the screen remains mounted, background return, process interruption and native alert callbacks are not covered. Undo is outside this pass. |
| Loading | Pending actions disable visible commands; the candidate also guards the callback synchronously. Native announcement/order of busy state remains pending. |
| Layout, adaptation, typography | Header menus are platform-owned, but alert wrapping, tablet anchoring, narrow windows and largest text sizes have not been inspected for these commands. |
| Appearance, localization | Destructive semantics are explicit. Locale translation, RTL, contrast and long names remain pending. |
| Targets, gestures, keyboard, accessibility | The native menu has a contextual accessibility label; explicit command controls are present. VoiceOver/TalkBack order, focus restoration, keyboard commands and physical reachability remain pending. The direct availability button is still a custom Pressable and needs native-adapter fit review. |
| Motion | No action-specific animation was introduced. Actual alert/menu motion and Reduce Motion behavior remain pending. |
| Search, privacy, notifications, media | These commands do not implement search or media selection, but their access checks and scope come from other layers. This review does not establish security-boundary coverage, push entry, photo interactions or background permission handling. |

Sources: AssetHeaderOverflow.ios.tsx, AssetOverflowMenu.tsx,
AssetLifecyclePresentation.ts, AssetDetailIdentitySection.tsx and
AssetDetailRouteScreen.tsx. Related edit/move loading controls and contained-item
search were located during discovery but are not included in these conclusions.

Validation for M86:63 remote asset-route checks pass, along with TypeScript and
mobile structural validation. The new callbacks were tested through the rendered
commands and confirmations, not by invoking private functions. This does not
prove mounted-but-blurred behavior or native menu state semantics.

## M110 — retained visit completion

The previously excluded blur/refocus path is now covered for checkout, return,
archive, restore and deletion. Ten success/failure cases plus three retained
confirmation cases failed first; focused visit/resource ownership now suppresses
old presentation and navigation. Three additional cases retain normal focused
success while preventing reuse of a completed confirmation. Pending ownership and
mutation-observer reconciliation remain intact. All91 focused tests, TypeScript and
structural checks pass remotely. Native alert, gesture, background and remount
behavior remain unverified; photo/undo completion is outside this change.

## Overflow S095 current source completion

S095 reviewed at f80a4b7e across all24 axes in the table above, alongside the
current Details route report. The iOS header installs UIBarButtonItem/UIMenu;
AssetOverflowMenu is the shared fallback/Android consumer. Do not transfer the
Sharing SwiftUI ellipsis bounds to the separate iOS header implementation.
History, lifecycle and permanent deletion remain separate command groups with
capability-derived availability and native destructive semantics for deletion.

There is no search, text editing or media acquisition inside overflow. The
containing Details route owns keyboard/content/media and query recovery; native
menu anchoring, target bounds, long names, focus and route-return behavior remain
unverified. Access-failure header cleanup remains an integration check, not a
proven pass. Existing112 Details checks support command ownership and presentation
behavior; they do not establish this menu's native geometry.
