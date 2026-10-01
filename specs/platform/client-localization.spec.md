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
rendered-copy gate or establish that unreviewed strings are acceptable.

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
