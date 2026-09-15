# Asset command review

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
