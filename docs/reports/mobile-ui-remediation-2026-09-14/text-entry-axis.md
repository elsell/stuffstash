# Text-entry ownership review

Source revision90bc977a, September15. `text-input-sites.csv` enumerates36 JSX
input/wrapper sites:28controlled,7native-seeded,1primitive forwarding site.
Included wrappers directly forward raw text values/events (OnboardingAddressInput,
OnboardingTextInput, ReturnNoteInput, and CustomizationLabeledInput). Compound
workflow/component invocations are not counted again as input sites. These
are source sites, not unique user fields: wrappers and their call sites overlap,
and dynamic forms render multiple fields from one site. A TypeScript generic
false positive was removed during inspection. Counts are not acceptance results.

The native controlled-address comparison loses characters on both devices;
Return notes lost/reordered characters on iPad. Native/uncontrolled address
comparisons pass. This is evidence of risk in the current runtime, not proof that
all controlled fields are broken. M65 is a Return-note candidate pending rerun.
No generic AppTextInput rewrite is justified without preserving external changes.

Current follow-up at c27a8eab: focused350563 completes8/10 iPad and7/10 phone.
Default-assisted SwiftUI passes both, while RN failures cross controlled and seeded
ordinary inputs. Full350549 phone reproduces Add name loss (`Nft name`) in the
navigation-stack case. A scoped iOS Add name candidate now uses SwiftUI TextField,
retaining ordinary assistance and the existing nameRevision reset ownership. It
does not change Android or other field families. The adapter test covers seed
retention, clear/resource remount, disabled edits and retry;25 related source tests,
TypeScript and structural checks pass on paul. Critic found no blocker. Focused
`add-draft` runs the four existing product journeys unchanged; native sizing,
appearance, keyboard, full-string entry and draft recovery remain acceptance gates.

Focused35058684319 at348e5fe8 completes2/4 phone Add journeys: stack entry and
the preconfigured-header route both preserve full text and pass. The latter
matches production's registered header. The old hidden-header diagnostic fails
before reaching the field; it is not evidence of current production typing loss.
The unfinished-tag journey now reaches the separate tag field and visibly retains
`Cing` instead of `Camping`. iPad completes3/4: both name journeys and the existing
tag journey pass; only the hidden-header pre-field diagnostic fails. That does not
erase the observed phone failure or imply all iPad fields are verified.
The inspected [phone rejected-save capture](evidence/phone-add-native-name-350586.png)
shows the complete name in both the field and rejection message, with keyboard
closed, commands visible and no clipping in this normal-size light state.

The scoped adapter is now named AddDraftNameField and also serves the inline tag
name. Successful staging advances its local reset revision; clear/restoration uses
the parent draft revision. Disclosure restores the persisted unfinished name.
The extended workflow test failed before implementation on paul, then25 related
tests, TypeScript and the mobile structural check passed. Assertions use persisted
draft values plus native seed ownership, rather than assuming a controlled field.
Critic found no blocker. This remains a native typing/layout candidate until the
unchanged tag journey passes on phone and iPad; no global text-input change is made.

## Reviewed task families and required ownership

| Family / source | Text task | External changes that must survive any fix | Next native acceptance |
| --- | --- | --- | --- |
| AddAssetScreen (five sites) | Name, description, parent/tag searches, new tag | Restore inventory/principal draft; reset after creation; parent selection replaces query; adding tag clears new name | Resolve Add loading first; full strings, restored draft, tag clear, rejected Save |
| AssetDetailSheets (six sites) | Edit name/description/tags; move destination/search | Opening a different asset or task seeds different values; choosing parent replaces query; adding tag clears draft | Edit/cancel/reopen, switch asset, move selection, full-string Save |
| CustomizationEditorFields / CustomizationEditorScreen (five sites) | New enum option; shared name/description/key | Add option clears draft; changing name derives stable key until manually edited; resource change loads another editor | Preserve generated/manual key transitions, validation focus, retry and reset |
| CustomizationCollectionScreen | Collection search | Scope/resource changes and search actions | Query typing, clearing, results, navigation return |
| ProviderProfileEditorScreens (two sites) | Secret replacement and prompt guidance | Successful credential save wipes secret; failure retains draft; owner/profile changes | Full input and retry using synthetic secrets only; masking and no disclosure |
| InventorySharingScreen | Invitee email | Inventory changes and successful invitation clear the form | Full email, permission denial, success clear, failure retains |
| VoiceConversationComposer | Typed request | Send clears composer; retry/history/context may replace it | Full message, submission, interrupted send and deliberate replacement |
| VoiceSessionSheetScreen (two sites) | Proposed item name and containing-location search | Command/plan identity changes; save/cancel name; parent selection closes search | Keep editing tied to command, full text, cancel and plan replacement |
| AssetContainedWorkspace | Search current contents | Asset scope and explicit clear/query updates | Type, clear, move between assets |
| AssetExpirationEditor | Search item types | Editor/session/type choice updates | Search then choose; reopen/reset |
| ExpirationField | Month-precision year | Precision conversion may seed year; Clear resets it; parent revision replaces editor | Enter year, convert precision, clear, preserve validation |
| ReminderTimingEditor | Custom advance days | Clean policy refresh updates days; dirty draft resists refresh; preset/custom changes | Numeric entry, refresh while dirty, Save/retry, keyboard dismissal |
| TagColorPicker | Hex value | Opening seeds current color; system spectrum selection changes hex; Clear resets it | Alternate typed and spectrum input; clear/cancel/apply |
| HomeReturnDetailsSheet (wrapper+site) | Optional note | New return session starts a new native field; errors retain draft | M65 full-string typing and failed-save retry on phone/iPad |
| Onboarding (five seeded sites) | Address and shared household/inventory fields | Step changes seed appropriate field; failure retains typed native state | Address button/Go already pass349270; other fields and full step transitions pending |
| AppTextInput | Native primitive forwarding | Caller defines controlled or native-owned semantics | No blanket behavioral substitution |

The source review follows actual change handlers and reset paths, not only JSX
`value` matches. Controlled does not mean inappropriate: color selection, derived
keys, precision conversions and composer resets need deliberate external updates.
A future shared input abstraction must distinguish user edits from explicit reset
or replacement commands. Dropping `value` everywhere would lose those behaviors.

## Native navigation search (separate path)

SearchScreen, InventoryMapScreen, BrowseFiltersScreen tags and TimeZonePicker use
NativeNavigationSearch. ExpirationWorkspaceScreen and ExpirationFiltersScreen set
native headerSearchBarOptions directly. These are excluded from the36 JSX sites
because their UIKit search fields are created by navigation. Their keyboard, query
clearing, blur/return, and scope state remain separate acceptance work; native
search cannot be inferred safe from the text-view comparison.

## Boundaries and next work

No new input corruption finding is asserted from source alone. This inventory
makes the outstanding keyboard/editing work explicit and guards against a blanket
fix that would discard legitimate programmatic updates. Runtime categories still
include rapid typing, paste, dictation/composition, autocorrection, selection/caret,
keyboard submit/dismissal, external hardware keyboard, denied/pending state,
resource change, and retained draft after retry. Inspect native evidence for each
family; do not mark keyboard cells complete from this inventory.

Critic review caught four omitted direct-wrapper calls; the inventory now includes
them. The family-level task review already covered their behavior.

## M74 — Add name field native editing candidate

Run349297 phone navigation-stack Add reached the form/header but typed Native draft
name as Nve draft name (screenshotA008EDF5-CE83-488A-80D8-8BEC399DC00F). The sheet
comparison failed before its field appeared. Keep these as separate findings.

Only iOS Name now uses a stable initial native value. The application still receives
edits for persistence/save/validation; restore/explicit clear/successful creation
advance the field lifetime. Metadata refresh and failed Save retain it. Android
keeps both its controlled editing and stable field key. Description and search/tag
fields remain unchanged because their reset semantics need separate review.

Six remote Add behavior checks, TypeScript and structural checks pass. Tests now
verify submitted/persisted names rather than relying on a controlled value prop;
restored seed and successful reset are checked separately. Critic found no blocker;
requested reset coverage and Android key scope were addressed. The unchanged native
full-string/rejected-save scenario remains the acceptance gate. This is a candidate,
not a claim that text corruption or Add sheet loading is resolved.

The CSV above is the baseline36-site inventory reviewed at90bc977a and committed at904684a1. This patch replaces its
Name site with AddAssetNameField (iOS seeded/Android controlled); use this delta when
reviewing the current tree rather than treating the baseline counts as current.

The added successful-reset case caught a candidate key collision with the expiration
editor. A name-specific key prefix corrected it; the final run has no duplicate-key
warnings and all six checks pass. This caught regression is not shipped native evidence.

### Ordinary keyboard comparison added after run349397

AddAssetNameField and ReturnNoteInput both already seed defaultValue through a
stable ref; neither supplies controlled value on iOS. Their ordinary-keyboard
failures therefore cannot be attributed solely to controlled JS value updates.
The runner's existing RN uncontrolled URL and SwiftUI URL inputs have different
keyboard and correction settings from those product fields.

Runner-only seeded single-line and multiline comparisons now use ordinary default
keyboard settings and check both native and application-observed full text.
Existing product scenarios and typing speed remain unchanged. Remote structural
checks pass; macOS compilation and execution are pending. This is diagnostic
coverage, not a fix or product acceptance.

An [older React Native controlled-input issue](https://github.com/facebook/react-native/issues/44157)
relates correction and JS synchronization, but it is marked fixed and does not
establish the cause here: this repository pins RN0.83.6 and the failing fields
are already seeded. No dependency or system keyboard preference was changed.

Critic identified the new bottom entries can be farther than one swipe from
the shared input slot; the test now uses a bounded reachability loop and retains
the hittable assertion before typing. Both fixture-preparation checks pass.

Run34947056524 iPad: ordinary single-line and multiline comparisons pass, while
controlled and seeded URL inputs still lose characters. All those RN comparisons
use the same thin AppTextInput wrapper. This does not establish a wrapper defect.
The new seeded URL/no-accessory comparison holds input settings and text constant
and removes only the extender before focus. The keyboard provider remains. Native
accessory attachment calls reloadInputViews in the pinned library; this is a
plausible variable to isolate, not a proven cause or reason to remove production UI.

### Run349725 phone address comparison could not reach the field

The system-address [final capture](evidence/phone-address-below-viewport-349725.png)
shows the fixture menu. Its hierarchy places the empty system field at y895.7–929.7
while the visible scroll viewport ends at y874. Tested-source36c45c3b line882 is the
pre-focus hittable assertion; the preceding loop always swipes down, away from a
field below the viewport. All four phone address comparisons failed at this line.
Those outcomes are not text-entry failures and do not contradict the earlier iPad
system/uncontrolled typing passes. Add/Home/Sharing malformed values are separate.

The address and ordinary comparison helpers now share bounded geometry-aware
reveal, scrolling in the appropriate direction until the whole input is in view.
Keyboard readiness, full-speed typing and native/application full-string checks
are unchanged. Native execution remains required; this is a test-procedure fix,
not a claim that product text corruption is resolved.

Both fixture-preparation checks, mobile TypeScript and structural checks pass on
paul. Critic found no blocker. The reveal helper checks vertical containment and
hittability for these full-width fixture inputs; it is not a general horizontal
clipping validator. Swift compilation and native interaction remain pending.

### Sharing email candidate M121

The iPad349789 capture retains a truncated email. Sharing now uses a mount-stable native
seed on iOS with lifetime reset for successful creation or scope change, while
Android retains controlled editing. Application state still drives submission and
validation. Twenty-two Sharing checks cover reset/retry semantics; the unchanged
native typing journey remains required. This supersedes the baseline controlled
classification for this one site, not the other input families.

M122 adds focused-route ownership to shared native search callbacks. Six consumers
were inspected: list/map Browse, Browse tags, Expiration selections, asset contents
and timezone choice. Full1,619-test validation passes; this does not establish
caller debounce/request cancellation or native return behavior. Expiration filter
selections now use the shared adapter; its results screen retains its separate
debounced hook with M120 focus handling.

### Focused Add run35060706395 — native tag adapter

At f7b09835, phone completes2/4 and iPad3/4. Both stack and
production-equivalent preconfigured-header Add recovery journeys pass; the older
hidden-header diagnostic still fails before field readiness. The full tag
disclosure/staging journey passes on iPad.

Phone stops before reaching the tag field: the immediate asset-name assertion
reads `T` after typing `Tent`. Its [final capture](evidence/phone-add-name-settled-350607.png)
and inspected final hierarchy both show `Tent`. This supports delayed observation
for this assertion, not an inference that all prior text failures were timing.
The test now waits at most5 seconds for exact `Tent`, then retains the equality
and all following draft/save checks. No paced typing, prefixes or disabled text
assistance. Phone tag acceptance remains unverified until that journey completes.

Phone job104685220242/artifact10433691877; iPad job104685220604/artifact10433966175.
Mobile structural check passes on paul; Swift/native execution of the corrected
assertion remains pending. Both original failures remain recorded.

### Native search enable transition

A shared adapter regression reproduced a blank native field when a mounted search
starts disabled and is later enabled with a retained, unchanged query. The prior
mount-only initialization had already run with no native field; its query-change
synchronizer saw no changed text. Enabling now seeds the current query, disabled
states receive no native writes, and native edit echoes still avoid text writeback.

The fake native command port regression fails before the correction and passes
afterward, including disable/change/re-enable and native typing echoes. Fifty
focused search/asset/location checks, TypeScript and structural checks pass on
paul; critic found no blocker. Consumer inspection includes Browse list, containment
Map, timezone, voice location and asset contents. Asset contents currently remounts
its search on enabled changes, so this is an adapter transition defect, not proof
of a previously shipped asset-contents failure. Native ref attachment timing and
M232's clear-collapse remain unverified. Logs: `/tmp/native-search-enable-*`.
Combined remote validation passes1,898 tests across296 files after this change
(`/tmp/native-search-enable-full.log`). This is source evidence, not native QA.


## Upstream comparison — September16

The installed React Native version is0.83.6. The historical
[cursor/prediction issue44157](https://github.com/react/react-native/issues/44157)
is marked fixed and describes controlled-only behavior on0.75.4. It is not a
confirmed explanation for this audit's failures, which also include a native-seeded
RN name field. The open [CJK composition proposal56082](https://github.com/react/react-native/pull/56082)
addresses marked-text preservation in Fabric; the current retained English/address
captures do not establish that marked-text corruption caused these failures.
Neither source justifies a blind dependency upgrade or broad native patch. Compare
against the actual0.83.6 input implementation and preserve external value/reset
ownership before adopting a correction. See the latest retained field/mirror values
and paced-versus-unpaced results in [phone351214](native-phone-351214.md).
