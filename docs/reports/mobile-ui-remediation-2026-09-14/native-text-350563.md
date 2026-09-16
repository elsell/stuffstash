# Focused text-entry run35056372549

Sourcefdbf30bf, focused workflow selection `text-entry`; this is not a full native
acceptance run. iPad job104667325488 completed8/10 passing. Phone104667325335
remains active at this checkpoint. Artifact10430144671 retains iPad evidence.

| Comparison | iPad result |
| --- | --- |
| Controlled assisted RN, no accessory | Fails: `Nve draft name` |
| Controlled assisted RN, normal accessory | Fails: `Nate draft nameiv` |
| Controlled RN without assistance | Pass |
| Uncontrolled RN single-line, multiline | Both pass |
| Uncontrolled RN without accessory/assistance | Both pass |
| Paced controlled/uncontrolled RN | Both pass |
| Default-assisted SwiftUI TextField | Pass |

Every case retains the same expected `Native draft name`; whole-string cases use
one typeText call. Native and observed application values are asserted. This run
narrows investigation toward assisted controlled RN editing and rules out the
accessory being necessary for these two failures. It does not establish all iOS
versions, physical typing, a particular dependency defect, or a universal cure.
Earlier uncontrolled failures remain evidence and must not be erased.

Do not disable assistance globally or slow production typing to match passing
cases. Before changing a shared adapter, inspect actual draft ownership and prove
programmatic updates, clears, resource replacement, refs, disabled state and
selection behavior remain correct. Add's iOS name already uses a revision-keyed
native draft; a blanket conversion would duplicate or break existing ownership.
Await phone comparison and current full-run Add/Edit journeys to prioritize the
production change. No code fix is claimed by this diagnostic.
