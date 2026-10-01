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

The required-checks and pre-commit `client-message-check` gate rejects embedded
copy directly rendered as JSX/Svelte text, display attributes, conditional labels
or fallback/template messages. It preserves protocol strings and variable-derived
user content. It does not prove the provenance of every variable or application
error; the remaining-file inventory and source review cover those boundaries.
