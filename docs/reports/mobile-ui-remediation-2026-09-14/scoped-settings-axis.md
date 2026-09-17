# Household and inventory settings — all 24 source axes

R034/R042, source53975079 plus M164. Reviewed route dispatch, ScopedSettingsScreens,
SettingsScreenState, SettingsList/styles, refresh notice and mounted settings
journeys. Category editors, notification settings and voice setup remain separate
surfaces; this does not claim their acceptance from the entry screens.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Identify the selected household/inventory and enter its management categories. Household definitions affect shared configuration; inventory rows identify the selected inventory. |
| Navigation | Household routes to fields/types/voice; inventory routes to sharing/reminders/tags/fields/types. These are substantial tasks, so navigation rows fit. No flat value picker is replaced by an unnecessary destination here. Native return remains pending. |
| Selection | No local value selection. Scope is selected from Home's inventory switcher. Entry rows do not silently change scope. |
| Modality | No modal introduced by either screen. Destinations own editing and dismissal. |
| Layout | Main/error/denied content uses scroll containers and shared settings insets. Loading is a centered view. Actual root-header inset, bottom reachability and long-name layout need native capture. |
| Adaptation | Shared row layout responds to font scale with stacked content. Normal-size narrow/iPad layout still requires runtime verification before enlarged-text work. |
| Typography | Scope name is a wrapping heading with household/inventory subtitle. Category labels and context use shared styles. No source line cap hides full scope text. |
| Appearance | Shared semantic palette and grouped settings rows. Contrast, reduced transparency and native bar composition remain runtime checks. |
| Localization | Scope names are user content; labels and interpolated Shared by text remain English. RTL row order and long translated copy are not verified. |
| Imagery | Small category symbols and disclosure chevrons provide navigation cues; labels carry meaning. There are no content photos. |
| Targets | Navigation rows declare minimum52-point height and44 width. Native Retry is shared. Actual hit regions and focus reachability remain open. |
| Gestures | All actions have visible tap controls. Vertical scroll is standard; no pull-to-refresh is added to these category indexes. |
| Keyboard | No text input. Hardware keyboard and assistive focus traversal remain unverified. |
| Accessibility | Rows have scope-specific names, button roles and destination hints. Denied state moves/announces accessibility focus. M164 adds labeled progress. Native focus timing and reading order remain open. |
| Motion | No custom transition/animation. Native navigation and progress must still be checked with reduced motion. |
| Content | Household shows3 categories when configure is allowed; inventory shows4 plus conditional Sharing. Lists are bounded and do not need pagination or search. |
| Search | No local search; category-specific collections own discovery/search. No additional persistent search field is justified here. |
| Loading | Scoped query initially blocks category rows. M164 identifies household/inventory loading using existing SettingsLoadingRow. Background ordinary errors keep usable context. |
| Recovery | First-load Retry uses native command; cached errors show SettingsRefreshNotice. Household denied has explanation. Retry pending/duplicate behavior in the shared refresh notice remains a follow-up review point, not certified by this pass. |
| Editing | No draft/mutation. Entering a category does not save changes. Destination editors retain their own draft semantics. |
| Privacy | Household configure gates all household categories; Sharing requires inventory share permission. Other inventory destinations may support viewing without editing. Query keys isolate session/scope and access errors remove cached data. Server authorization remains separate. |
| Notifications | Inventory row is labeled Notifications with Your reminders as context, and navigates to notification settings. No system notification delivery/tap logic is handled here. |
| Media | No acquisition/playback operation. Voice setup configures a separate capability; it does not start recording here. |
| Lifecycle | Shared scoped model reloads with session/inventory key changes. Routes record settings-level selection through observability. Current native scope-switch/back/denied transitions remain pending. |

Both new labeled-loading cases failed before M164 and now pass, including
replacement by the correct category content. Existing household denial and
settings-retry tests remain in the mounted suite. No native status is promoted.
