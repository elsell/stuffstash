# Full-run text-entry traces — 35247151136

Tested merge f7b3f995dd2ef0e9f26bba2f48d6dfd759075bbd. Range extraction retained
11 phone and12 iPad RN traces from artifacts10511454857/10512433387, using about
1.6MB of downloads instead of the combined4.3GB archives. The raw traces are in
[phone evidence](evidence/phone-input-events-352471.json) and
[iPad evidence](evidence/ipad-input-events-352471.json). All buffers report zero
dropped events. The missing phone paced-controlled trace is not treated as a pass.

Every captured RN comparison receives the complete expected key sequence, including
all corrupted values. Examples:

| Comparison | Phone final recorded JS value | iPad final recorded JS value |
| --- | --- | --- |
| Controlled address | `h/example.invalid` | `hinvalid` |
| Controlled name | `Native t namedraf` | `N` |
| Controlled name without accessory | `Nve draft name` | `Native dr nameaft` |
| Seeded ordinary name | `Nt name` | `Native draft name` |

The iPad controlled-name trace contains change `N` at count1, React commit `N`,
then every remaining key in `ative draft name`, with no further change or
selection event. This corroborates the earlier [key-event run](native-text-entry-351529.md);
it is not the first evidence of complete key receipt.
The trace's final value is JavaScript state, not an independent reading of the
backing field; native acceptance assertions remain recorded in the full-run logs.

## Current investigation boundary

The installed RN0.83.6 delegate emits key events before editing the backing field.
The fixture does not configure maxLength. Therefore receipt of keys does not
establish acceptance or retention of the edits. Native value mutation, selection
restoration and suppressed change delivery remain candidates. A later snapshot
or slower typing is not a correction.

The [provider-free run](native-text-entry-351564.md) already reproduced corruption
with verified provider omission on iPad. Removing the keyboard provider or merely
hiding its accessory is not justified. Likewise, assistance-disabled and seeded
fields have failed in prior samples; neither is a general solution.

Next native instrumentation should distinguish the backing field's text and
selection before/after delegate edits from framework state writeback, without
altering the typed string or production behavior. Focused35803226783 remains the
current independent observation of the existing fixture; it does not add this
lower-level instrumentation. No production fix is claimed by this report.
