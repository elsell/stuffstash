# Definition and tag collection routes

R029/R032/R037/R040/R046: household asset types/fields, inventory asset types/
fields/tags. Each route delegates to CustomizationCollectionRoute and the shared
CustomizationCollectionScreen. Reviewed atd458d896 with M156/M157 candidates,
including CustomizationCollectionQuery, useCustomizationReads, access policy,
native header/search adapters and mounted collection tests.

Search is an in-place refinement; Add opens a distinct editor. M156 uses existing
native header adapters, following the user's accepted compact-search preference.
This is a project pattern decision, not a claim that Apple requires icon-only
search in every collection. Relevant Apple topics are
[search fields](https://developer.apple.com/design/human-interface-guidelines/search-fields)
and [toolbars](https://developer.apple.com/design/human-interface-guidelines/toolbars);
the text-only fetch returned JavaScript placeholders during this pass, so no new
specific HIG claim is inferred from those responses.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Browse definitions, distinguish household inheritance, refine and create when permitted. Tag rows retain compact names/color semantics. |
| Navigation | Kind/scope routes choose the correct collection and editor target. M156 places Add in the native header; actual navigation/return remains native acceptance. |
| Selection | Active/Archived uses the existing native segmented adapter. Opening a definition navigates; inherited details remain read-only. |
| Modality | Collection is a pushed destination; no new sheet is added for search. Editor interaction is reviewed separately. |
| Layout | M156 removes custom search/Add scroll chrome. Content, lifecycle control and notices retain shared insets; native header/keyboard geometry remains pending. |
| Adaptation | Shared rows wrap and flex. Phone/iPad native header fit, long names and dense lists are not established by source checks. |
| Typography | Row names are primary, definition metadata secondary; tag names are not duplicated as metadata. Normal long names precede enlarged-text work. |
| Appearance | Rows/notices use palette tokens; search/Add inherit platform appearance. Measured contrast and both native appearances remain open. |
| Localization | Name search normalizes local case and trims query. English labels and user-provided household names remain; RTL/translation unverified. |
| Imagery | Tags reserve a color indicator slot and expose non-color names. Definitions use disclosure chevrons; no media thumbnails. |
| Targets | Row minimums are 52/58 points in source; Add/search targets are native. Native hit bounds remain a separate acceptance check. |
| Gestures | Explicit search/Add/open/Retry; pull-to-refresh uses usePullRefresh. Native scrolling and navigation return remain unverified for these routes. |
| Keyboard | Shared native search owns focus and search submission; closing/clearing resets filtering. Body uses automatic keyboard insets/dismissal. Native cancellation/return remains pending. |
| Accessibility | Rows have names including inheritance/tag color. Native search and Add are labeled. Reading order, header announcements and color contrast require runtime review. |
| Motion | No bespoke collection animation. Native navigation/search and reduced-motion behavior remain pending. |
| Content | Real query traverses bounded pages and sorts names. Inherited/local sections remain distinct; partial loading never claims complete results. |
| Search | M156 replaces custom controlled text input with native navigation search. Mounted tests filter on input, retain query through refresh and restore rows on cancel. No server search is claimed. |
| Loading | Initial loading removes header commands; lifecycle transition retains known rows with progress. Explicit pulls own the indicator, independent of background fetching. |
| Recovery | Initial failure has Retry; stale refresh retains known rows; partial lists show an honest warning. Empty and no-match states differ. Native recovery reachability remains open. |
| Editing | Add only for current authorized active collections. Lifecycle operations occur in detail; collection itself does not mutate definitions. |
| Privacy | Existing scope/read policy gates rows. M157 corrects stale Add permission and deactivates retained action handlers on revocation; mounted positive/revoked checks complement existing boundary tests. |
| Notifications | N/A: these collections do not schedule or consume notifications. |
| Media | N/A: no acquisition, upload or playback. Color is a detail field, not collection media. |
| Lifecycle | Resource generations ignore stale loads; denial removes rows. Shared search rejects hidden callbacks; native state restoration and platform interruption remain pending. |

Three search-adapter regressions failed before migration. The permission-revocation
case also failed before reading the current permission snapshot. All56 focused
collection/search/header tests pass remotely, as do TypeScript, structural checks
and fixture preparation checks. A synthetic native collection journey now tests
Add, search, exact filtering and cancellation. It has not yet run in a native
build; no visual or gesture pass is claimed.

Combined remote validation: all1,704 mobile tests across270 files, TypeScript,
mobile structural rules and both fixture-preparation tests pass
(`/tmp/mobile-collections-batch.log`). Critic review found no production blocker;
the native test now waits for the excluded row to disappear before claiming
filtering completed. Its Add check verifies dispatch, not editor navigation.
