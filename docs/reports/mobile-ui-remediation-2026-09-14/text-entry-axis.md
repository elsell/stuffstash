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
