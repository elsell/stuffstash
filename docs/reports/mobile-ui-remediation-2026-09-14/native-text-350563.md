# Focused text-entry run35056372549

Sourcefdbf30bf, focused workflow selection `text-entry`; this is not a full native
acceptance run. iPad job104667325488 completed8/10 passing. Phone104667325335
completed7/10. Artifacts10430144671 and10430943560 retain their evidence.

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
Phone results prevent attributing all failures to controlled assisted editing:
ordinary controlled RN passes, but no-accessory controlled reports `Nve draft name`,
no-assistance controlled reports `Ne draft name`, and ordinary uncontrolled
single-line reports `Nive draft name`. Multiline, uncontrolled no-accessory and
no-assistance, both paced cases, and default-assisted SwiftUI pass. The inspected
uncontrolled final screenshot retains `Nive draft name` in both field and observed
application text; this is not merely an early partial assertion. See
`evidence/phone-uncontrolled-text-350563.png`.

The scoped Add name candidate uses the successful SwiftUI path with the existing
revision-owned seed. Twenty-five related tests and static checks pass on paul;
critic found no blocker. These checks establish adapter/draft behavior, not native
fidelity. Unchanged product Add journeys must rerun before accepting the candidate.
