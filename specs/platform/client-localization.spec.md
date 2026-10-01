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
