# Move routes and selection surfaces — all 24 axes

R013 Move, R014 Move Here, S134 destination selection, S135 creation and S136
source selection. Reviewed at bba8c1e3 with M196 candidate. Sources:
AssetNativeActionSheetScreens, AssetDetailSheets, AssetDetailMovePresentation,
AssetNativeSheetOptions, useParentCandidates and both route adapters.

| Axis | Current source evidence and remaining acceptance |
| --- | --- |
| Task | Move chooses where the current asset goes; Move Here chooses an existing asset for the current destination. Creation is a separate explicit command within Move, now native via M196. |
| Navigation | Success records completion for the target Details workspace and returns with router.back. Direct-entry return, both route stacks and focus restoration need native checks. |
| Selection | Searchable hierarchical candidates justify this selection view. Parent rows show title/kind/path and Selected; the two creation kinds use NativeChoicePicker in place. Move excludes self/noncontainers; Move Here excludes self/current children. Server validation still decides cycle legality. |
| Modality | Native form sheets have grabbers and bounded detents. Idle Cancel closes; operation lock prevents dismissal while creation/move is pending. Native drag and repeated Back remain acceptance work. |
| Layout | Current title, help, placement, input, candidates and preview are inside one ScrollView. NativeSheetActions sits below it. Earlier fixed-content/custom-footer notes are historical. Actual safe-area and keyboard fit remain pending. |
| Adaptation | Flexible content and wrapping rows exist; detents differ for Move/Move Here. Normal phone/iPad and narrow-window captures still required. |
| Typography | Move separates short heading from long asset subject; path text wraps. Creation labels include the entered name. Native long-label and normal-size path wrapping remain unverified. |
| Appearance | Palette drives selection and grouping; action footer, kind choice and creation command use shared native adapters. Light/dark and disabled contrast remain native checks. |
| Localization | Copy is English; Move Here preview still uses a text arrow. RTL, long translations and locale-sensitive name comparison remain review gaps. |
| Imagery | No essential photo information: rows use name/kind/path. Similar-name disambiguation depends on paths. Icons/photos are not introduced by M196. |
| Targets | Parent rows have64-point minimum height and explicit selected/disabled states. Native commands own action targets. Actual bounds, long-row clipping and competing sheet gestures remain unverified. |
| Gestures | Explicit selection, Move and Cancel complement system sheet gestures. Busy operation blocks removal. Native scrolling versus sheet expansion must be tested. |
| Keyboard | Shared AppTextInput, keyboard-aware scrolling and container keyboard avoidance are used; footer declares container ownership. Combined native inset behavior and restored text entry remain unverified. |
| Accessibility | Selection rows expose role/state; creation now has one complete action label. Check VoiceOver grouping of kind/path, selection announcements and return focus. |
| Motion | No custom task animation. System sheet and keyboard animation, reduced motion and interrupted dismissal remain runtime checks. |
| Content | Move previews current/proposed path; Move Here previews chosen asset/target. Creation selects the new destination without moving the original asset yet. Cancelling after creation leaves the created destination; destructive rollback is not implied. |
| Search | Trimmed queries debounce250ms, use scoped keys and hide unsettled results. Creation requires known candidates and avoids an exact same-kind/name/parent match. It is not a global uniqueness guarantee. |
| Loading | Suggestion progress occupies results; creation/move freezes inputs and commands. “Creating destination…” distinguishes creation from moving. Unknown candidates never authorize creation. |
| Recovery | Suggestion retry preserves query and stays inside the scroll region. Move/create failures retain draft and release operation lock. Native alert/keyboard/action accessibility remain pending. |
| Editing | Selection is provisional until Move. Creation mutates independently then selects its result. Synchronous operation ownership prevents double creation or simultaneous Move and freezes draft changes. |
| Privacy | Scoped asset/candidate queries and injected commands are used. UI visibility does not establish server authorization or containment validation; backend security coverage is separate. |
| Notifications | No notification-specific controls; app interruption during a pending task is an unresolved runtime scenario. |
| Media | N/A: these tasks acquire or play no media. |
| Lifecycle | Operation completion is focus-owned; stale completion does not alert or navigate after leaving/returning. Resource replacement and native keyboard/dismissal transitions remain runtime acceptance. |

Remote M196 validation:82 related tests plus TypeScript and structural checks pass
on paul (`/tmp/move-create-command-green.log`). The revised named-command cases
failed first and still exercise duplicate creation, blocked Move, disabled kind,
locked input and late result isolation. Critic found no blocker.

Before M196, the accumulated bba8c1e3 tree passed all1,815 mobile tests/284 files,
TypeScript and structural checks on paul (`/tmp/mobile-audit-bba8c1e3-full.log`).
Checksum comparison matched tracked mobile source/fixtures/config; React act
warnings remain. Neither result establishes native acceptance or release readiness.
