# Client Localization Spec

## Scope and status

This closes G6 only when mobile and web presentation strings and plural messages
are catalog-backed and representative expansion/RTL checks pass. Infrastructure
alone is not completion. English is the initial shipping translation; additional
production languages require reviewed catalogs and runtime acceptance. Protocol
values, IDs, user content and provider output must never be translated as UI copy.

## Boundary

A dependency-free `packages/localization` package owns message catalogs and pure
formatting. Both clients consume it through their own presentation adapter. It
must not depend on API transport, React, Svelte, storage or a mutable global locale.
Use stable, contextual message keys and whole messages with named placeholders;
never assemble translated sentences from fragments. Reusable components, routes
and presentation helpers move copy out of implementation files into catalogs.
Application/domain protocol enumerations retain their stable wire values.

Resolve formatting locale from the client runtime's Intl locale unless explicitly
injected for verification. Resolve message language independently: missing languages
fall back to English, and English fallback messages use English plural categories,
not the device language's categories. Use Intl.PluralRules for cardinal selection;
counts must be finite numbers. Date/time, number and list helpers use Intl. Missing
required interpolation arguments fail visibly during development/tests. Never
interpret message or interpolation values as HTML, code or navigation targets.

Each translator instance is immutable and independent. Catalog loading must not
change another instance's locale (including simultaneous browser/SSR contexts).
The initial apps select locale on startup; a runtime language preference UI is not
part of this capability. Later translations can be installed without component edits.

## Verification locales

Support explicit `en-XA` expanded/accented and `ar-XB` RTL pseudolocales for
verification. Only catalog text is transformed; user item names, URLs, IDs and
interpolated values remain intact. Preserve placeholders before interpolation.
Expose text direction so web/native hosts can choose their platform direction
without forcing a production device-wide RTL reload. No claim of native RTL
acceptance may be made from a pure formatter or browser test.

### Authentication recovery messages

Mobile authentication-required errors must carry cataloged text for session
refresh failure, unusable sign-in tokens and missing required refresh tokens.
Some command screens can show the typed error message before navigation returns
to sign-in; onboarding's separate fallback does not cover those consumers.
Preserve error types, token validation, storage clearing, refresh coalescing and
session-generation isolation. Verify the three failure paths through the existing
session ports under a pseudolocale, alongside the existing security regressions.

## Migration and enforcement

Migrate reusable controls and everyday workflows before administrative surfaces.
Keep a concise remaining-file inventory until no embedded presentation messages
remain; distinguish technical constants and verbatim user/server content. CI must
check catalog validity and migrated surfaces for accidental new embedded copy.
Import structural checks must inspect each declaration independently: a localization
value import before a React Native type-only import must remain allowed, while a
type-only import before a forbidden runtime import must not hide the violation.

Critical tests cover plural zero/one/many, English fallback under a non-English
locale, parameter safety, independent translators and pseudolocales. Representative
browser and native workflows verify expanded labels, search, approval and return.
Do not create a separate test for every extracted label.

### Archive review counts

Both clients must render restore-review counts using shared locale-aware plural
messages: item/tag/photo/file totals, custom-type/field totals, and complete
omission/remapping notices. Zero, one and many must select the appropriate form.
Keep summary ordering in catalog templates; never concatenate an English noun to
a formatted number. A shared presentation formatter accepts only preview counts,
not transport/domain objects or user titles. Critical tests cover singular and
mixed totals, zero/many notices and expanded-locale output. This changes wording,
not archive selection, validation or approval.

## References

- [ECMA-402 internationalization API](https://402.ecma-international.org/)
- [Expo localization guide](https://docs.expo.dev/guides/localization/): Intl uses
  the device locale when no explicit locale is supplied.


## Semantic state and localized copy

Import count tiles carry a stable metric identity. Visibility, icon selection and
navigation to issues or imported records use that identity, never English words
inside labels or action descriptions. Zero blocking-issue counts remain visible
under every locale. The display count and noun may form a metric tile, but full
summary sentences use catalog plural forms. Move creation labels choose complete
messages by location/container kind; conversational plan counts use full plural
messages instead of supplying English nouns as parameters.

Import issue grouping and guidance use stable cause categories derived from source
codes and unchanged legacy diagnostic fields. They must not inspect translated
cause labels. Unknown source diagnostics stay verbatim; recognized causes and
client explanations are catalog-backed. Group identity remains stable when only
the display language changes.

The required-checks and pre-commit `client-message-check` gate installs the pinned
workspace dependencies before loading its compiler parsers, then rejects embedded
copy directly rendered as JSX/Svelte text, display attributes, conditional labels
or fallback/template messages, including responsive table labels rendered through
CSS data-cell-label attributes. It preserves protocol strings and variable-derived
user content. It does not prove the provenance of every variable or application
error; the remaining-file inventory and source review cover those boundaries.

Conversation review intent uses an explicit approve/cancel value while a decision
is pending. Progress wording is presentation only; translating it must not change
whether the user sees saving or cancellation. An approved server plan remains a
save even if an earlier client cancellation label or intent is present.

Native locale evidence must exercise production controls, not an isolated text
sample. A runner-only pseudolocale Add workflow verifies native header actions,
exact user-entered text, rejected-save draft retention, error/header clearance and
return. Expected accessibility labels come from the same reviewed catalog rather
than a second translation implementation. Screenshots and element hierarchy are
retained for visual inspection. RTL pseudotext evidence alone does not certify
native navigation mirroring or physical assistive behavior.

Settings and conversation-case validation use complete catalog messages, including
required/optional character and UTF-8 byte limits. Invitation action labels,
confirmations and accessible names localize independently of action IDs, status,
URLs and permission decisions; email addresses remain verbatim parameters.

Import and sharing notices, recovery instructions and confirmation copy must use
catalog messages. Source diagnostics remain verbatim; known relationship/status
values get presentation labels without changing authorization values. Import dates
and file sizes follow the client formatting locale, not a fixed US locale.

Import history decides whether records changed from numeric counts, never from the
localized “no records changed” label. A skipped-only run must not gain an empty
change summary when the display language changes.

The final presentation pass includes accessibility announcements, native choice
summaries, notification recovery, onboarding guidance, and administrative status
maps. Native module identifiers, keyboard keys, time-zone IDs, technical exceptions
and semantic state values are not messages. Styling or section visibility must
use explicit state flags instead of comparing translated labels.

Parent suggestions derive root/parent status from containment identities, not an
English location label. Application-produced validation, recovery and status copy
uses catalog messages; onboarding's existing diagnostic matching stays stable
while its user-facing messages translate at the presentation boundary.

## Native Intl compatibility

Hermes does not provide every Intl API available in browser/Node verification.
The mobile presentation adapter installs pinned FormatJS compatibility modules
before constructing its translator: intl-getcanonicallocales 3.2.12, intl-locale
5.3.11, intl-pluralrules 6.3.15, and intl-listformat 8.3.15. Install only missing
APIs and register English data for the initial shipping message language; future
production translations must register corresponding plural/list data. Preserve
native number/date/collation formatters. This runtime compatibility dependency
belongs to mobile, not the dependency-free shared message package. Verify startup,
plural selection, and list formatting with optional Intl APIs absent, followed
by the existing macOS native workflow. Node success alone is not native acceptance.

The locale prerequisite uses the September 12 release (and its exact supported-values dependency), satisfying the fourteen-day supply-chain review window without an age exemption.

The remaining-copy review must follow variable-derived labels and command notices,
not just JSX/Svelte literals. Catalog-backed announcements include entity-link
context and duplicate position, disconnected review recovery, upload partial
success, move/save results, onboarding placeholders and parent-search summaries.
Use complete plural messages for counts. Preserve user/provider text as parameters
and leave protocol matching, developer diagnostics, CSS, key names and fixture data
untranslated. Classify those exclusions explicitly in the remaining-file inventory.

Keyboard focus and element lookup must use owned element references or stable IDs,
never translated accessible names. Browse Map jump selection/Escape returns focus
to its bound search input regardless of locale.

Save-with-parent and photo-upload outcomes use complete catalog sentences for each
context, with cardinal forms for upload counts. Expiration labels select complete
state/precision messages, including the month-end qualifier; map child counts use
catalog plural messages. Preserve safe provider reasons and user titles verbatim.

Import headings/range summaries, upload constraints, reminder summaries and
identity-recovery notices are complete catalog messages. Byte displays retain
binary scaling and the existing B/KB/MB convention while formatting numeric
values with the selected locale. Product labels never determine identity or flow.

## Native RTL audit integrity

The RTL audit must verify native layout direction, not only Unicode-isolated
English copy. Runner-only XCTest launches use the pinned React Native
`RCTI18nUtil_forceRTL` launch default before bridge initialization and verify the
actual `I18nManager.isRTL` value exposed by the isolated fixture. Add/recovery
assertions include mirrored leading Close/trailing Save geometry and retained
user text. Non-RTL runs explicitly launch without forced RTL. These settings
remain in the audit harness; production startup and device preferences are unchanged.
A run that fails direction or geometry cannot count as RTL acceptance.

Reference: [React Native I18nManager](https://reactnative.dev/docs/i18nmanager).

Literal-message inventory is a triage aid, not a count of missing translations.
Retain reviewed classifications alongside generated candidates, identifying exact
strings when only part of a file was reviewed and caller evidence where errors
are replaced with localized guidance. Keep protocol diagnostics and test/benchmark
fixtures separate from product copy. A classification does not suppress the
rendered-copy gate or establish that unreviewed strings are acceptable. Reviewed onboarding/authentication diagnostics map through
`onboardingError`; archive task/transfer diagnostics map through the mobile
archive report function or web archive failure function; ordinary voice transport
and session errors map through `buildFailedVoiceRealtimeState`. Retain those
caller mappings in the inventory instead of translating internal exceptions.
This source review does not establish native recovery or assistive acceptance.

Step-progress navigation accessibility labels use complete catalog sentences for
current, completed and upcoming steps, with step names/descriptions interpolated.
Do not concatenate English instructions with translated state labels or lowercase
translated labels to build a sentence. Keep navigation reachability unchanged.

Native system surfaces are part of client copy coverage: invitation share text
and title, export share-sheet title, and Android notification-channel display
name use the catalog. Preserve invitation URLs and inventory names verbatim,
copy-to-clipboard behavior, MIME/UTI values and stable notification channel IDs.

Mobile repository presentation labels also use catalog messages: fallback location
and update labels, known search-match field names, and complete dated-update
messages. Preserve server titles, descriptions, unknown field identifiers and
existing date formatting; translating a display label must not change API values.

## Diagnostics fallback copy

Settings diagnostics must resolve missing server/version labels through the client
catalog. Actual configured URLs and version identifiers remain verbatim; the
authentication-mode enum remains a protocol value.

## Command validation shown in the UI

Invitation input validation and provider connection
test failures are presentation messages when their callers display Error.message.
They must use catalog messages just as labels do. Preserve the existing exception
types, validation order and English wording; do not translate server/provider data
or protocol identifiers. Legacy errors already mapped by a presentation boundary
remain stable inputs to that mapping.

### Voice preview query failures

The voice preview must render cataloged recovery copy for directory/transport
failures instead of displaying arbitrary exception messages. Preserve the typed
selected-inventory-unavailable guidance, which is already cataloged. Internal
cursor/page-limit and server messages remain diagnostics, not untranslated UI.

### Customization keys and inventory query guidance

Customization key rules remain pure domain predicates. Their explanatory copy
belongs to the catalog and is shared by inline editor validation and command
rejections. Home query guidance rendered to users is also
catalog-backed. Preserve existing English wording and validation semantics.

### Option-object copy enforcement

Copy enforcement must inspect display properties such as label/title in nested
component option objects and script-defined option arrays in both JSX and Svelte.
Stable value/id/state properties, CSS strings, user content and catalog calls
remain exempt. Include an adversarial regression where translated direct props
coexist with untranslated nested choices; direct-attribute coverage alone is
insufficient.

### Adapter-generated recovery copy

Mobile timeout and provider-onboarding failures are application-authored recovery
messages surfaced by existing error presenters. They must use catalog messages;
timeout/caller-abort classification and tenant checks remain unchanged. Preserve
external server diagnostics and user content rather than translating arbitrary
error strings. Native expansion evidence identifies its source revision, runtime,
workflow and limitations independently from physical acceptance.

Retry policy must recognize a typed network timeout independently of localized
message text. A timeout permits at most the existing single read retry; caller
cancellation never becomes retryable merely because it interrupts the transport.

### Representative native search and proposal acceptance

The pseudolocale native suite reuses production search/retry/selection/return and
pending-proposal close/reopen/reset-protection workflows alongside Add/recovery.
Generate expected accessibility labels from the production catalog and formatter,
including interpolated location names and the localized inventory-root fallback.
Preserve fixture names, user search input and system-owned keyboard labels.
Expanded and RTL runs remain separate evidence and neither substitutes for actual
VoiceOver/TalkBack use or physical integrations.

Native acceptance observers must re-query replaceable search fields while waiting
for user-entered values. A stale XCTest element must not be treated as loss of
input when the current hierarchy retains it. Diagnostic header probes must expose
React state separately from native title state, distinguishing missed interaction
from presentation updates without changing production behavior on inconclusive evidence.

Browse's inventory-name fallback must use the existing inventory catalog label
when no selected inventory record is available. A real inventory name remains
verbatim; this fallback does not change navigation or selection behavior.

Household setup recovery must present the partial-success explanation and recovery
guidance as a complete catalog message. Do not render the application exception
message directly. Preserve the created household, inventory draft, and inventory-only
retry so localization cannot cause duplicate household creation.

Unavailable checkout, return, return-details and undo commands produce cataloged
guidance because Home and Details display these command errors. Keep unsupported
command identifiers as developer diagnostics and preserve action dispatch and IDs.

Inventory discovery recovery must use catalog guidance. Selecting an unavailable
inventory must use the same typed unavailable error as reading a removed selection,
without changing the previously selected inventory. Repeated continuation cursors
and the bounded page limit must fail with cataloged recovery, preserving their
existing request bounds. Cancellation remains control flow, not recovery copy.

### Attachment upload recovery

Mobile attachment upload recovery must use complete catalog messages for three
failure groups: invalid target or method, failed upload response, and missing
local file content. Invalid targets and methods must fail before loading or
calling native upload code; they must not trigger JSON fallback or completion.
Keep HTTPS/local-development policy, permitted methods and successful upload
behavior unchanged. Missing file guidance must ask the user to choose it again.
Keep native file transfer injectable at the transport adapter so rejection and
completion behavior can be checked without replacing native modules.

### Svelte collection expressions

The rendered-copy gate must inspect display properties inside Svelte each-block
collection expressions, including conditional arrays and template-valued headings.
Do not interpret collection IDs or arbitrary source values as display copy. Browse
List/Map labels and tenant/inventory grouping headings for asset types and custom
fields must use complete catalog messages, preserving user names as parameters.

### Complete paged-read recovery

Map hierarchy, inventory tag selection and Home checked-out reads must show
cataloged retry guidance when their shared pagination guard rejects a missing or
repeated continuation or reaches its page limit. Preserve the guard's existing
request bounds and initial-cursor tracking. Shared consumers, including photo
and search reads, retain their existing fallback and cancellation behavior.

### Browse and location recovery

List and search pagination that cannot safely continue (missing or repeated
continuation cursor) must stop at the existing guard and show cataloged retry
guidance, without exposing cursor diagnostics. A location missing from the
selected inventory must show cataloged unavailability guidance. Preserve selected
inventory scope, user-authored names, cancellation and existing request bounds.

### Closed-state messages

Workspace creation failures, invitation copy/share failure titles, and active/
archived settings loading labels use complete messages for each closed state.
Do not interpolate internal kind, action or lifecycle identifiers into translated
sentences. Preserve English recovery meaning, draft retention, privacy of transport
errors, and command behavior. User-authored names remain verbatim parameters.

### Browser photo upload recovery

Storage HTTP rejection and transport failure must use cataloged recovery text even
when the exception is explicitly safe for presentation. Preserve direct-upload
failure: neither case may silently retry through JSON or complete metadata.
Critical tests cover both causes, presentation through the safe-error boundary and
expanded-locale output. Keep technical target-selection and invalid-invitation
errors distinct from this deliberately user-visible upload error.

### Asset region recovery sentences

Details contents/photo failure, loading and retry messages must be complete
catalog entries for each region. Region identifiers are control-flow values,
not user text to interpolate. Preserve independent region retries and disabled
retry controls while a request is running. Verify rendered messages in the
expanded pseudolocale so untranslated region words cannot hide inside a
translated sentence.

### Catalog-owned navigation labels and destination fallback

Expiration filter page titles and search prompts use catalog entries, never
capitalized route identifiers. History's selected filter accessibility label uses
the same cataloged option label as its picker. The Add destination root fallback
is cataloged; selected destination paths and user-authored unresolved names remain
verbatim. Preserve filter draft staging, page transitions and destination choices.

### Interpolated literal enforcement

The rendered-copy gate must inspect literal interpolation values passed to the
project's `t` function in TypeScript, JSX and Svelte. Catch direct English
values, conditional/fallback values and values wrapped in `String`; a catalog
template does not translate those values. Preserve message keys, variable
references containing user text, nested catalog calls, numeric formatting and
unrelated protocol/style values. This is bounded syntax checking, not proof of
variable provenance or full client migration.

The first interpolation-gate migration covers mobile voice-placement fallbacks,
asset lifecycle notices, notification read actions and deferred entity labels;
web containment permission guidance, notification actions, expiration summaries
and conversation evaluation/activation summaries. Preserve user names, dates,
durations, operation identifiers and selection state. Product-owned fallback
words and sentence fragments must come from catalogs; read/unread actions and
deferred entity labels use complete alternative messages.

### Web domain-value presentation

Render asset lifecycle badges, customization scope/boolean type labels, and
conversation case/outcome/operation summaries through typed catalog mappings.
Reuse the same mappings in expected and observed summaries, including forbidden
and executed operation lists. Keep wire enums, fixture IDs, authored names and
technical verdict codes unchanged; localization must not change evaluator inputs.

### Web item creation recovery

Creation and post-save refresh failures use cataloged recovery for ordinary
exceptions through safeWorkspaceErrorMessage. Preserve explicitly safe server
validation and user names. Saved items, created parents and tags must survive
failures without duplicate creation. Unavailable tag creation uses an explicitly
safe cataloged message. Verify initial failure, partial success and unavailable
tag capability through the real workflow; preserve pseudo-localization.

### Mobile read and photo-selection recovery boundaries

Home initial-load and refresh, location initial-load, and voice-plan photo-selection
failures must not display ordinary exception messages. Use their cataloged fallback
guidance. Preserve deliberately catalog-owned failures (empty workspace, missing
tenant/location, camera permission and unsupported photo format) through a typed
error carrying a catalog key, never an arbitrary safe-text flag. Retain retry,
existing content, photo drafts and departed-visit suppression. Verify mounted
recovery and the actual photo provider's permission failure.

### Mobile list and asset-action recovery

Apply the catalog-only error boundary to inventory/location lists, asset-details
commands and Home return/details/undo actions. Preserve cataloged command and
upload precondition guidance. Lifecycle recovery may recognize the existing
active-children or archived-parent diagnostic category, but never interpolate
the diagnostic itself into visible text. Use complete cataloged action guidance,
retain authored asset names, and preserve retry, draft retention and visit scope.

### Mobile editor, settings and sharing recovery

Use catalog-only recovery for Add/edit/move submission, account/provider settings
commands, and invitation creation/copy/share/cancellation. Preserve cataloged
validation, explicit missing-link guidance, and field/destination/photo drafts.
Provider required-field validation carries a catalog key rather than arbitrary
text. Ordinary exceptions must not appear in inline errors, notices or alerts.
Keep duplicate-command guards, one-time invitation handling and retry behavior.

### Voice readiness labels and raw-error regression guard

Both provider-readiness errors and voice failure presentation must name missing
capabilities with the same cataloged labels used by provider settings. Retain
wire capability IDs in typed data and reject unknown values from visible lists.
The copy gate must reject direct ordinary-Error message/fallback conditionals in
rendered text, display properties, and returned presentation helpers. Typed
catalog recovery and diagnostic-only categorization remain permitted. This
syntactic guard complements caller review; it is not data-flow verification.
