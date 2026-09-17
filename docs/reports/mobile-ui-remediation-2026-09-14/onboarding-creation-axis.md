# Household and inventory creation — all 24 source axes

S060 household setup, S061 first inventory and S062 partial recovery at71183281
plus M206. Reviewed OnboardingScreen/Presentation, OnboardingCommand, HouseholdSetup,
profile persistence and existing application/mounted tests. See onboarding-axis.md
for platform guidance and browser/server-entry scope.

| Axis | Source conclusion and remaining acceptance |
| --- | --- |
| Task | Name a household and first inventory as prerequisite setup; after partial success, finish only the inventory. No optional tutorial steps added. |
| Navigation | State changes in one form, with retained connection profile. Completed setup hands off to the authenticated app. Start Over signs out and clears setup. |
| Selection | Freeform names require text entry, not a menu. Existing usable inventory discovery bypasses unnecessary creation. |
| Modality | Creation remains in context; browser authentication belongs to the connection step. No confirmation before requested creation. |
| Layout | Full-width scroll container, centered maximum-width form, safe areas and keyboard avoidance. Native command migration requires new reachability checks. |
| Adaptation | One column with wrapping text; tablet margin scrolling was separately tested before M206. This is not new-control native acceptance. |
| Typography | Clearly labeled names, heading, defaults and recovery copy. Long name/error and enlarged-text runtime acceptance remain open. |
| Appearance | M206 replaces custom action buttons with shared native primary/standard commands. Existing semantic form palette remains. Actual light/dark native appearance pending. |
| Localization | English copy/default inventory name; user names preserved and trimmed at command boundary. Non-Latin names/RTL remain unverified. |
| Imagery | Shared local brand mark; no remote media dependency in setup. |
| Targets | Native command adapter replaces bespoke buttons; help remains a44-point disclosure. Actual phone/tablet targets and keyboard overlap require retest. |
| Gestures | Explicit submit/reset/help; scroll dismisses keyboard. No gesture-only completion. |
| Keyboard | Household/inventory multiline-free fields use native-owned seeded text; keyboard Go shares required-value validation and duplicate lock. Native typing correctness remains a separate acceptance gate. |
| Accessibility | Visible labels, named commands, step-heading focus and alert error text. M206 preserves visible command text while separately exposing progress. Native speech/order remains unverified. |
| Motion | No custom animated onboarding sequence. Native command/keyboard transitions and reduced-motion behavior pending. |
| Content | First inventory has an editable default. Partial error explains household completion and the remaining task. No implementation identifiers in recovery UI. |
| Search | No list-search task in the naming form. Automatic discovery uses authorized services. |
| Loading | Synchronous submission lock, disabled inputs/commands, separately labeled progress. Command label remains visible after M206. |
| Recovery | Definitive rejection permits retry; ambiguous writes reconcile before repeating. Household success is retained when inventory fails. Error preserves the entered inventory name. |
| Editing | Missing required values disable progress with explanation; invalid nonempty server input is validated on submit. Start Over clears drafts; unmounted completion cannot navigate. |
| Privacy | Browser sign-in and authorized API ports precede creation. Read-only household access cannot create inventories. No new authentication/authorization behavior in M206. |
| Notifications | No notification permission requested merely to finish account setup. |
| Media | No camera/library/audio/file access. |
| Lifecycle | Generation checks prevent superseded command completion; saved profile permits relaunch discovery. Ambiguous recovery state is session-local; process-death/OS browser interruption require separate native acceptance. |

One RED case established disappearing command text during pending sign-in. The
native migration passes40 tests across five onboarding/invitation files plus
TypeScript and structural checks on paul. Critic found no production blocker;
its test cleanup removes reliance on a void button handler as a completion promise.
Native phone/iPad tests must rerun with M206. Earlier successful onboarding runs
verify the prior controls, not this implementation or full text-entry behavior.
