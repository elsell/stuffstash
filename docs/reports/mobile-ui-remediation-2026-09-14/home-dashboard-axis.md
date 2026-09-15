# Home dashboard — all 24 source axes

R002, source81e91f74 plus M160. Inspected HomeRoute, HomeScreen, its styles and
header, HomeDashboardQuery, scoped query and pull-refresh hooks, notification and
expiration entry composition, and return-task ownership. Header, summary and
return-task reports retain their own scope; this review covers their composition.

Apple's [loading](https://developer.apple.com/design/human-interface-guidelines/loading)
and [scroll view](https://developer.apple.com/design/human-interface-guidelines/scroll-views)
topics are relevant references. Their current HTML returned JavaScript-only content
and DocC retrieval was unavailable during this pass; no new claim about Apple's
wording is made. M160 is an engineering ownership defect demonstrated by mounted
behavior, independently of those references.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Dashboard supports resuming work, identifying expiration and returning borrowed items. Bounded summaries lead to complete lists. No new task or compulsory setup is introduced. |
| Navigation | Header enters Add/inbox/account/switcher. Recent and checkout rows open detail; section actions open full collections. Mounted tests cover destinations. Native tab/back/deep-link return remains incomplete. |
| Selection | Inventory choice belongs to the scoped switcher; summaries navigate rather than maintaining a hidden local selection. Flat value selection is absent from this route. |
| Modality | Only Return creates a bounded details task, through the shared task provider and native route. Save/Cancel ownership is separately reviewed. Ordinary browsing does not introduce a modal. |
| Layout | Dashboard ScrollView requests automatic insets; shell owns left/right safe areas. Header actions remain outside scroll content. Centered loading/error states are fixed Views: long error text and short windows need runtime reachability checks, not an assumed pass. |
| Adaptation | Flexible row cards and width-aware header share the layout. Phone/iPad native fixtures have remaining failures; full tab/accessory composition is not established by the isolated header fixture. |
| Typography | Section headings and long item names use shared text styles. Normal-size long content remains part of acceptance. Enlarged-text remediation is deferred per user sequencing, not certified. |
| Appearance | Semantic palette feeds shell, text, cards and spinner. Shared native header policy owns material. Actual transparency, light/dark contrast and reduced-transparency behavior remain native checks. |
| Localization | Shared card date formatting is used; route copy is English. Error content comes from query errors and may vary in length. RTL, translated copy and date rollover need separate acceptance. |
| Imagery | AssetCard owns photos/placeholders and expiration symbols. Dashboard introduces no additional image acquisition or custom icon grammar. Missing-image behavior retains shared media acceptance gaps. |
| Targets | Section links declare minimum44-point dimensions; header actions are native. Full native run350129 measured header AX frames36points high. This failed geometry check is not proof that UIKit's actual hit region is36points; retain the unresolved runtime distinction. |
| Gestures | Explicit taps navigate and issue Return; native vertical scrolling/pull refresh are used. No swipe-only command. Spinner source ownership is tested; actual navigation-return geometry still needs native acceptance. |
| Keyboard | Dashboard has no text input; scroll uses shared dismiss policy. Return details and global voice own input. Keyboard return/focus is not proven by dashboard tests. |
| Accessibility | Section headings have header roles; asset, location, section and action controls have descriptive labels. Loading includes visible text. Actual VoiceOver order, announcements and native target behavior remain open. |
| Motion | No custom dashboard animation. Native transitions, scrolling and progress animation require reduced-motion runtime checks; no essential information is animation-only in source. |
| Content | Recent and checkout summaries each cap visible rows at3 from up to10 queried items. Expiration composes an independently queried summary. Explicit full-list actions preserve access to longer lists. |
| Search | No local search field. Full Browse/expiration destinations own search/filter state; absence here is an intentional summary pattern, not missing functionality. |
| Loading | First load shows labeled progress; cached data remains during ordinary refresh. Explicit pull alone owns the native indicator. Return reconciliation is independent. M160 scopes delayed pull-error feedback to its original visit/resource. |
| Recovery | Initial failure has native Retry. Access failures hide cached dashboard data; ordinary refresh failure retains it. Current pull failure shows feedback; old-visit failure is silent after M160. Background query errors with retained data have no route notice; freshness affordance is a remaining design consideration. |
| Editing | Dashboard itself has no draft. Return enters a details task with Save/undo-aware Cancel and permission gating. Its inventory-keyed owner rejects old editor callbacks; see separate return review/tests for exact guarantees. |
| Privacy | Query keys include server session, tenant and inventory; access failures suppress data. Add/Return follow create/edit permissions. These client gates do not replace API authorization tests. |
| Notifications | Bell composition uses selected settings scope and a keyed NotificationBell. OS notification taps belong to the system-entry surface, not this source review. Physical notification behavior is not newly verified. |
| Media | Dashboard displays images only; acquisition is in destinations. Global voice accessory belongs to the tab shell. Camera, audio and interruptions require their own runtime coverage. |
| Lifecycle | Resource replacement remounts dashboard return ownership; shared header uses current committed callbacks. Pull hook clears its spinner on blur, and M160 also retires its feedback. Background date change and full cold/warm/tab journeys remain native gates. |

M160 adds two failing-before-fix mounted navigation cases and a positive current
visit failure case. All28 Home cases and2 pull-hook cases pass remotely on paul;
TypeScript/structural results are recorded in the checkpoint. This is not a native
Home acceptance pass. Android runtime coverage is still unavailable.
