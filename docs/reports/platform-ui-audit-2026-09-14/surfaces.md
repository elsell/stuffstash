# Surface coverage and assessment

Revision `8392eeb4cedbd7e327c344f4e6b5a78a2fff4851`. The
[source inventory](source-inventory.md) enumerates 62 route/layout files and 313
production UI component/style sources. Web's catch-all route additionally expands
into the modes and task states below. Mobile settings collections share their
implementation across household and inventory routes.

“Source review” means the interaction structure was inspected; it does not mean all
branches were executed. “Inventory” means the entrypoint and composition were
accounted for without detailed behavioral verification. No row claims a runtime
pass. Sources referenced in findings carry the strongest evidence.

| Surface family | Entrypoints and nested tasks | Source/design assessment | Outstanding acceptance |
| --- | --- | --- | --- |
| Mobile shell | Root stack; Home/Browse native tabs; voice accessory; keyboards | Source review: real native tabs and native-stack foundation. Custom content remains inside native containers. F04/F09. | Native insets, top/bottom edges, keyboard, tab return, hardware Back on Android |
| Home | Dashboard, expiration counts/rows, recently changed, checked out, inventory menu, Add/Notifications/Profile | Source review: native actions and gesture-owned Home refresh now present; do not re-report old screenshots as current defects. | Header fit at large text, long inventory name, refresh/return and permission-dependent actions |
| Browse List | Search, peer List/Map switch, summary, grid, sort/filter, pagination, detail return | Source review: native compact search and shared header; filter friction F01. Grid layout is not inherently un-native. | Search entry/cancel, selected filters, keyboard, empty/error, saved scroll position |
| Containment Map | Horizontal hierarchy, column selection, breadcrumbs, search/path expansion, details | Source review: custom hierarchy has a product reason; Reduce Motion support present. Geographic Maps HIG N/A. | Deep paths, duplicate names, narrow/wide, focus order, nested gestures and top clearance |
| Browse filters | Overview, type, status, availability, tags, sort, expiration handoff | Detailed source review: F01; retain draft/apply and guard scope changes. | Menu draft/cancel/commit; tags search; expiration handoff context; busy cancellation |
| Expiration | Soon/Expired/All dates, month groups, search, detail return, refresh/more | Source review: system segments with menu fallback for constrained layouts; F05/F09. | Date boundaries, long labels, search, refresh ownership, no results and stale data |
| Expiration filters | Kind, availability, types, multi-tags, locations, date range | Detailed source review: F01 and F10; long searchable choices justified. | Duplicate location names, all/none, range errors, keyboard, draft dismissal |
| Assets/locations lists | `/assets`, `/locations/[locationId]`, nested asset detail | Source review: ordinary object navigation appropriate; F05. `LocationsScreen` also inventoried, reachability uncertain. | Background refresh, no photos, large text, list/detail return |
| Asset detail | Photos, identity, fields, tags, containment, overflow, checkout/return, archive/restore/delete | Source review: content and contextual commands appropriate; action wrappers shared. | Permission changes, failed commands, destructive recovery, dynamic content, nested taps |
| Add | Title, kind, parent lookup/quick create, types, fields, tags, photos, expiration, draft restore | Source review: meaningful editor; mobile draft store exists; F02/F03/F04. | Draft restore/scope change, keyboard, partial creation, permission denial, image acquisition |
| Edit / Move / Move here | Native sheets; details, tags/date; destination search, hierarchy and selection | Source review: bounded tasks justify sheets; edit has dirty-close prompt and disabled gesture dismissal. | All cancellation methods, partial save/failure, deep parent selection, async return |
| History/checkouts | Asset history list, change detail, revert, checkout history | Source review: separate detail justified by content and revert decision. | Revert confirmation/result, unavailable event, long values, scroll/focus return |
| Photo/files | Source chooser, camera/library, upload progress/retry, gallery/full-screen viewer | Source review: system acquisition plus custom viewer; separate accessibility assessment required. | Limited/denied access, canceled acquisition, upload failure, zoom/swipe/explicit close, delete |
| Inventory switcher | Inventories/households internal modes | Detailed source review: F07; hierarchy can justify sheet. | Many inventories, large text, explicit close, selection failure, current context |
| Onboarding/auth | Server, browser sign-in, household/first inventory, retry/start over | Source review: necessary self-host setup, browser auth, focus change and error states. | Interrupted callback, wrong server, keyboard, returning account, pending invitation |
| Invitation acceptance | Link sign-in handoff, review, accept/dismiss, account mismatch, expired link | Source review: explicit task and recovery states; not a simple picker. | Cold/warm links, wrong account, expired/revoked invitation, VoiceOver focus |
| Notification inbox | All/unread, open item, mark read/unread/all, settings | Source review: native segmentation and value-rich rows; custom toolbar F04. | Cold/warm push tap, vanished item, scope mismatch, unread persistence, accessibility |
| Reminder settings | Permission toggle, default/type overrides, presets/custom days, timezone | Source review: switches and searchable timezone fit; preset/editor refinement optional, F04. | Denied/revoked permission, settings handoff, custom cancel, inherited state, saving failure |
| Settings root/account/appearance/server/about/diagnostics | Personal preferences and connection/identity information | Source review: category navigation legitimate; Appearance selection is a possible simplification, not an automatic violation. | Sign-out/server change confirmation, long values, current setting accessibility, enlarged text |
| Household/inventory settings | Tags, fields, asset types: list/new/detail/edit/archive/restore/delete; inherited/denied states | Source review: entity editors justify navigation, small field selectors do not need custom groups (F02). | Dirty guard, inherited/read-only, duplicate names, lifecycle actions, native color picker |
| Mobile voice experience | Entry accessory, typed input, recording, transcript, response, plan preview/edit/approval, progress/recovery | Source review of composition and controls; no new live conversational evaluation. Native sheet is appropriate for bounded interaction. | Mic denial, interruption, stop/cancel, typed fallback, announcement order, third-party data-use clarity |
| Mobile voice administration | Stage selection, provider list/add/detail, credentials, prompt guidance, test/activation | Source review: descriptions/configuration justify destinations; not equivalent to three-option filter. | Save failure, credential handling, denied access, long prompts, keyboard/focus |
| Web shell/search/context | Desktop sidebar, mobile nav, context/account menus/sheets, global search suggestions | Source review: browser primitives, links, combobox keyboard behavior and focus restoration present. | Browser Back, new tab, keyboard/zoom, empty suggestions, responsive overlays, screen reader |
| Web Home/Browse/Map/location/detail | Modes `home`, `browse`, `location`, `asset`; search/filter/grid/hierarchy/contained content | Source review of composition and shared patterns; preserve web conventions. | Narrow/wide, large lists, tabs vs links, hierarchy keyboard, loading/error/empty |
| Web expiration/inbox/reminders | Expiration mode, filter dialog, date inputs, inbox and personal settings | Source review: in-place Select already exists; bounded multi-tag region. | Dialog keyboard/focus, long options, many tags, dates/locale, no matches and retries |
| Web Add/edit/lifecycle/media | Add sheet; edit/move/move-here/checkout/return/archive/restore/delete; attachment actions | Detailed Add review F08; other action sheets use shared guarded patterns, not blanket loss claim. | Draft preservation, pending close, partial save, uploads, return focus and browser navigation |
| Web sharing/settings/customization | Grants/invitations, account, household/inventory collections, activity, administration | Source review: custom field editor has dirty handling; web Select/checkbox/group patterns remain appropriate. | Permission denial, role selection, invite cancel/expire/delete, keyboard, dirty navigation |
| Web import | Source choice/configuration, preview/issues/plan, confirm/start, history, run detail, records/timeline, cancellation | Source review of source setup, preview, history and detail: labeled credential/file inputs, explicit preview/start, progress and cancel states. A long task with meaningful stages; do not compress into a menu. | Runtime walkthrough pending: keyboard/focus through stages, failures, stop/resume, large reports |
| Web conversation administration | Workflow/case lists, revision editors, fixtures/expectations, test setup/results/comparison/activation | Source review of workflow/case editors, run setup/detail/activation: labeled Select controls, validation/status, explicit discard/replacement and revision comparison. Destinations justified; full interactive QA pending. | Runtime walkthrough pending: validation, keyboard, draft close, revision compare, errors/access |
| Documentation/public entry | Starlight sidebar/search/theme, landing/product/setup/configuration/help, invitation web fallback | Source review of shell/config/theme plus page inventory. Treat as documentation, not iOS chrome. | Browser keyboard/search, code copy, narrow/zoom, dark contrast, link destinations |
| Android resolution | Shared mobile routes plus native Compose menus/header/footer and fallback controls | Source review: native adapters exist, but generic choice fallback remains custom (F02). No emulator run. | TalkBack, system Back, edge-to-edge, keyboard, 48dp targets, native selection behavior |
| iPad/adaptive contexts | Declared tablet support across all mobile families | Scope established; F09. No evidence sufficient to certify adaptation. | Split window, rotation, keyboard/pointer, menu anchoring, accessibility sizes |

## State and input coverage required for each relevant family

The source survey considered empty, loading, denied, error, ready, and busy/draft
branches where present. Their runtime matrix remains pending: successful primary
path; unavailable data; large/duplicate/long content; keyboard and focus; cancel and
return; slow/failing work; light/dark; large text; reduced motion/transparency;
VoiceOver/TalkBack; browser zoom and semantic interaction; supported tablet widths.
Use synthetic inventories rather than mutating household data for audit evidence.

## What this audit does not infer

No unimplemented optional HIG integration is a missing requirement. No raw style or
component count proves a defect. No source test proves the current native rendering.
Import/conversation administration have source/control-flow coverage; documentation
has shell/theme and page inventory coverage. Their runtime walkthroughs remain pending. Backend, CLI,
and MCP contracts were not UI-audited; their human-visible effects are considered
through the surfaces that expose them.
